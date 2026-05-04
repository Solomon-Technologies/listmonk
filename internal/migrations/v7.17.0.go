package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V7_17_0: multi-tenant fork — Rule27 Design split.
//
// Adds schema-level `company_id` tenancy across every tenant-scoped table
// so a single Listmonk instance can host fully isolated workspaces. Two
// initial tenants:
//
//	id=1  Solomon Technologies (consolidates AnilTX, Auldrom, Solomon Tech,
//	      Byte Arch, Skulptor — future per-brand split is supported by the
//	      schema but deferred)
//	id=2  Rule27 Design (data backfilled from rule27-q2-* lists, Q2 expose
//	      campaigns, and the dedicated warming addresses/campaigns/senders)
//
// What this migration does (in order):
//
//  1. companies table + seed.
//  2. company_id INT NOT NULL DEFAULT 1 + FK + index added to every
//     tenant-scoped table (core, drips, ab, automations, scoring, CRM,
//     warming, segments, webhooks, users).
//  3. Backfill: every Rule27 row gets company_id=2 (lists 17,18,19,20,24;
//     campaigns 16,17,18; warming senders 11,12; warming campaigns 11,12;
//     all dependent rows by FK chain).
//  4. Subscriber duplication: 200 subscribers that exist on BOTH Rule27 and
//     Solomon lists get a duplicate row under company_id=2 so each tenant
//     owns its own copy. Their Rule27 list memberships are re-pointed to
//     the duplicate; Solomon-side membership of the original is left alone.
//  5. Subscriber email uniqueness swap: drop global UNIQUE(email) and
//     idx_subs_email, add UNIQUE(LOWER(email), company_id).
//  6. Other globally-unique tenant-scoped columns get per-company swaps:
//     templates.is_default, campaigns.archive_slug, warming_addresses.email,
//     warming_senders.email, roles.name.
//  7. Refactor messengers from JSON-in-settings to a dedicated `messengers`
//     table with FK to companies. Existing JSON entries are parsed and
//     inserted: `email-resend-solomontech` → company_id=1, anything matching
//     `*rule27*` → company_id=2, everything else → company_id=1.
//  8. Seed Rule27 admin roles. (Users themselves are seeded with disabled
//     status + password_login=false; Alchemy enables and sets passwords via
//     the admin UI after the migration runs.)
//
// All steps run inside a single transaction; partial failure rolls back.
//
// Handler-level enforcement of company_id is gated by the
// `app.enforce_company_isolation` config flag (off by default after this
// migration). The schema is in place either way; flipping the flag turns on
// the WHERE company_id = $user.company_id filters and the cross-tenant 404
// asserts. See cmd/handlers.go and security.md for the runtime model.
func V7_17_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.17.0 migration: multi-tenant fork (Rule27 split) ...")

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op if commit succeeds

	// --------------------------------------------------------------------
	// Step 1: companies table + seed.
	// --------------------------------------------------------------------
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS companies (
			id          SERIAL PRIMARY KEY,
			name        TEXT NOT NULL,
			slug        TEXT NOT NULL,
			created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_slug ON companies(slug);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_name ON companies(name);
	`); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO companies (id, name, slug) VALUES
			(1, 'Solomon Technologies', 'solomon'),
			(2, 'Rule27 Design',        'rule27')
		ON CONFLICT (id) DO NOTHING;

		-- Reseat the sequence so subsequent inserts don't collide with the
		-- explicit IDs above.
		SELECT setval('companies_id_seq', GREATEST((SELECT MAX(id) FROM companies), 2));
	`); err != nil {
		return err
	}

	// --------------------------------------------------------------------
	// Step 2: company_id column + FK + index on every tenant-scoped table.
	//
	// All columns default to 1 (Solomon) so no existing row violates the
	// NOT NULL constraint. Rule27 rows are re-tagged in step 3.
	//
	// Pure join tables (subscriber_lists, campaign_lists, campaign_media,
	// ab_test_assignments, ab_test_variants, drip_steps, drip_enrollments,
	// drip_send_log, automation_*, score_log, webhook_log, warming_send_log,
	// contact_activities, conversions, link_clicks, campaign_send_log,
	// campaign_views) inherit company_id via their parent FK chain — no
	// own column. Filtering through the parent is sufficient and avoids
	// drift.
	// --------------------------------------------------------------------
	tenantTables := []string{
		// Core
		"lists", "subscribers", "campaigns", "templates", "media",
		"bounces", "users",
		// Solomon-fork: segments, webhooks
		"segments", "webhooks",
		// Drips (parent only — steps/enrollments/send_log inherit)
		"drip_campaigns",
		// A/B testing (parent only)
		"ab_tests",
		// Automations (parent only)
		"automations",
		// Scoring (parent only)
		"scoring_rules",
		// CRM
		"deals", "conversion_goals",
		// Warming
		"warming_addresses", "warming_senders", "warming_campaigns",
		"warming_templates",
		// Auth roles — each tenant has its own role definitions
		"roles",
	}
	for _, t := range tenantTables {
		if _, err := tx.Exec(`
			ALTER TABLE ` + t + ` ADD COLUMN IF NOT EXISTS company_id INT NOT NULL DEFAULT 1;
		`); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			DO $$ BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM information_schema.table_constraints
					WHERE table_name = '` + t + `' AND constraint_name = 'fk_` + t + `_company'
				) THEN
					ALTER TABLE ` + t + ` ADD CONSTRAINT fk_` + t + `_company
						FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE RESTRICT;
				END IF;
			END $$;
		`); err != nil {
			return err
		}
		if _, err := tx.Exec(`
			CREATE INDEX IF NOT EXISTS idx_` + t + `_company ON ` + t + `(company_id);
		`); err != nil {
			return err
		}
	}

	// --------------------------------------------------------------------
	// Step 3: backfill Rule27 records.
	//
	// Identifiers locked at audit time (2026-05-03):
	//   Lists       17 (rule27-q2-eligible), 18 (rule27-q2-A), 19 (-B),
	//               20 (-C), 24 (rule27-q2-A-contacted)
	//   Campaigns   16, 17, 18 (q2-expose-A/B/C; messenger LIKE '%rule27%')
	//   Warming sn  11, 12 (Rule27 Design)
	//   Warming cp  11, 12 (active)
	//
	// IDs are stable in prod. If they ever drift, a future migration can
	// re-classify by name/messenger pattern.
	// --------------------------------------------------------------------
	if _, err := tx.Exec(`
		-- Lists
		UPDATE lists SET company_id = 2
			WHERE id IN (17, 18, 19, 20, 24) AND company_id != 2;

		-- Campaigns: by id AND by messenger pattern (defense-in-depth)
		UPDATE campaigns SET company_id = 2
			WHERE (id IN (16, 17, 18) OR messenger ILIKE '%rule27%')
			  AND company_id != 2;

		-- Warming senders + campaigns (active campaigns 11, 12 use senders 11, 12)
		UPDATE warming_senders SET company_id = 2
			WHERE id IN (11, 12) AND company_id != 2;
		UPDATE warming_campaigns SET company_id = 2
			WHERE id IN (11, 12) AND company_id != 2;

		-- Warming addresses linked to Rule27 senders/campaigns: defer manual
		-- review. Default DEFAULT 1 keeps shared addresses on Solomon side.
	`); err != nil {
		return err
	}

	// --------------------------------------------------------------------
	// Step 3.5: backfill subscribers that exist ONLY on Rule27 lists.
	//
	// These ~6,715 subs (per audit) have no Solomon list membership, so
	// they don't need duplication — they're cleanly Rule27's. Just re-tag.
	//
	// Cross-brand subs (on BOTH Rule27 and Solomon lists) are NOT touched
	// here; step 5 handles them via duplication so each tenant owns its
	// own copy.
	// --------------------------------------------------------------------
	if _, err := tx.Exec(`
		UPDATE subscribers SET company_id = 2
		WHERE company_id != 2
		  AND id IN (
		      SELECT subscriber_id FROM subscriber_lists
		      WHERE list_id IN (17, 18, 19, 20, 24)
		  )
		  AND id NOT IN (
		      SELECT subscriber_id FROM subscriber_lists
		      WHERE list_id NOT IN (17, 18, 19, 20, 24)
		  );
	`); err != nil {
		return err
	}

	// --------------------------------------------------------------------
	// Step 4: subscriber email uniqueness — global → per-company.
	//
	// IMPORTANT ORDERING: this MUST run before step 5 (subscriber
	// duplication). The original UNIQUE(email) constraint blocks any
	// INSERT of a row with an email that already exists, so duplicating
	// an overlap subscriber under company_id=2 fails until we drop the
	// global constraint and replace it with the per-company one.
	//
	// We swap:
	//   - column-level UNIQUE on email (auto-named subscribers_email_key)
	//   - functional case-insensitive idx_subs_email on LOWER(email)
	// for a single per-company functional unique index:
	//   - idx_subs_email_company on (LOWER(email), company_id)
	// --------------------------------------------------------------------
	if _, err := tx.Exec(`
		ALTER TABLE subscribers DROP CONSTRAINT IF EXISTS subscribers_email_key;
		DROP INDEX IF EXISTS idx_subs_email;
		CREATE UNIQUE INDEX IF NOT EXISTS idx_subs_email_company
			ON subscribers (LOWER(email), company_id);
	`); err != nil {
		return err
	}

	// --------------------------------------------------------------------
	// Step 5: subscriber duplication for cross-brand overlap.
	//
	// 200 subscribers (per audit) have membership on BOTH a Rule27 list
	// (17, 18, 19, 20, 24) and a Solomon list. Per Alchemy's decision, each
	// such subscriber gets a fresh row under company_id=2 so both tenants
	// own their own record.
	//
	// Then re-point those subscribers' Rule27 list memberships to the new
	// duplicates (so Rule27 sends to its copy, Solomon sends to the
	// original). Solomon-side memberships of the original are untouched.
	//
	// Idempotent via NOT EXISTS guard — re-running won't double-duplicate.
	// --------------------------------------------------------------------
	if _, err := tx.Exec(`
		WITH overlap AS (
			SELECT DISTINCT s.id      AS orig_id,
			                s.email,
			                s.name,
			                s.attribs,
			                s.status,
			                s.company,
			                s.phone,
			                s.lifecycle_stage
			FROM subscribers s
			JOIN subscriber_lists sl_r ON sl_r.subscriber_id = s.id
			                          AND sl_r.list_id IN (17, 18, 19, 20, 24)
			JOIN subscriber_lists sl_s ON sl_s.subscriber_id = s.id
			                          AND sl_s.list_id NOT IN (17, 18, 19, 20, 24)
			WHERE NOT EXISTS (
				-- Skip subs that already have a company_id=2 twin (idempotent guard)
				SELECT 1 FROM subscribers s2
				WHERE LOWER(s2.email) = LOWER(s.email) AND s2.company_id = 2
			)
		)
		INSERT INTO subscribers (uuid, email, name, attribs, status, company, phone, lifecycle_stage, company_id)
		SELECT gen_random_uuid(), email, name, attribs, status, company, phone, lifecycle_stage, 2
		FROM overlap;
	`); err != nil {
		return err
	}

	// Re-point Rule27 list memberships from the original (company_id=1) sub
	// to the new duplicate (company_id=2). The original keeps its Solomon
	// memberships only.
	if _, err := tx.Exec(`
		WITH twins AS (
			SELECT s_old.id AS old_id, s_new.id AS new_id
			FROM subscribers s_old
			JOIN subscribers s_new
			  ON LOWER(s_new.email) = LOWER(s_old.email)
			 AND s_new.company_id = 2
			 AND s_old.company_id = 1
		)
		UPDATE subscriber_lists sl
			SET subscriber_id = twins.new_id
			FROM twins
			WHERE sl.subscriber_id = twins.old_id
			  AND sl.list_id IN (17, 18, 19, 20, 24);
	`); err != nil {
		return err
	}

	// --------------------------------------------------------------------
	// Step 6: other tenant-scoped uniqueness swaps.
	// --------------------------------------------------------------------
	if _, err := tx.Exec(`
		-- templates.is_default: was global "only one default at a time".
		-- Now: one default per company. Index name has to differ since the
		-- old one was anonymous (CREATE UNIQUE INDEX ON templates ...).
		DO $$
		DECLARE old_idx text;
		BEGIN
			SELECT indexname INTO old_idx
			FROM pg_indexes
			WHERE tablename = 'templates'
			  AND indexdef ILIKE '%is_default%true%'
			  AND indexname NOT LIKE 'idx_templates_default_company';
			IF old_idx IS NOT NULL THEN
				EXECUTE format('DROP INDEX IF EXISTS %I', old_idx);
			END IF;
		END $$;
		CREATE UNIQUE INDEX IF NOT EXISTS idx_templates_default_company
			ON templates (company_id, type) WHERE is_default = true;

		-- campaigns.archive_slug: was globally unique. Now per-company.
		ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS campaigns_archive_slug_key;
		CREATE UNIQUE INDEX IF NOT EXISTS idx_campaigns_archive_slug_company
			ON campaigns (company_id, archive_slug) WHERE archive_slug IS NOT NULL;

		-- warming_addresses.email: scope to company.
		ALTER TABLE warming_addresses DROP CONSTRAINT IF EXISTS warming_addresses_email_key;
		CREATE UNIQUE INDEX IF NOT EXISTS idx_warming_addresses_email_company
			ON warming_addresses (LOWER(email), company_id);

		-- warming_senders.email: scope to company.
		ALTER TABLE warming_senders DROP CONSTRAINT IF EXISTS warming_senders_email_key;
		CREATE UNIQUE INDEX IF NOT EXISTS idx_warming_senders_email_company
			ON warming_senders (LOWER(email), company_id);

		-- roles.name: scope to company so each tenant can have its own
		-- "Super Admin" / "Operational Admin" role names without colliding
		-- with the legacy global ones.
		DROP INDEX IF EXISTS idx_roles_name;
		CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_name_company
			ON roles (type, name, company_id) WHERE name IS NOT NULL;
	`); err != nil {
		return err
	}

	// --------------------------------------------------------------------
	// Step 7: messengers refactor.
	//
	// Currently messengers live as a JSONB array under
	// settings.value WHERE key='messengers'. To enforce company isolation
	// at the DB layer, we pull them into a dedicated table with a
	// company_id FK.
	//
	// Tagging rule for the backfill:
	//   - Anything whose name matches '%rule27%' (case-insensitive) goes
	//     to company_id=2.
	//   - Everything else goes to company_id=1 (Solomon default).
	//
	// The settings JSON entry is left in place for backward compatibility
	// during the dual-mode rollout. Once `app.enforce_company_isolation`
	// is flipped on and verified, a future migration can clear it.
	// --------------------------------------------------------------------
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS messengers (
			id          SERIAL PRIMARY KEY,
			company_id  INT NOT NULL REFERENCES companies(id) ON DELETE RESTRICT,
			name        TEXT NOT NULL,
			type        TEXT NOT NULL DEFAULT 'smtp',
			config      JSONB NOT NULL DEFAULT '{}',
			is_default  BOOLEAN NOT NULL DEFAULT false,
			is_enabled  BOOLEAN NOT NULL DEFAULT true,
			created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE (company_id, name)
		);
		CREATE INDEX IF NOT EXISTS idx_messengers_company ON messengers(company_id);

		-- Backfill from settings.messengers JSON. Each element gets
		-- company_id=2 if its name matches rule27, else 1.
		INSERT INTO messengers (company_id, name, type, config)
		SELECT
			CASE WHEN (elem->>'name') ILIKE '%rule27%' THEN 2 ELSE 1 END AS company_id,
			elem->>'name'                                                AS name,
			COALESCE(elem->>'type', 'smtp')                              AS type,
			elem                                                         AS config
		FROM settings s
		CROSS JOIN LATERAL jsonb_array_elements(s.value) AS elem
		WHERE s.key = 'messengers'
		  AND jsonb_typeof(s.value) = 'array'
		  AND elem ? 'name'
		ON CONFLICT (company_id, name) DO NOTHING;
	`); err != nil {
		return err
	}

	// --------------------------------------------------------------------
	// Step 8: Rule27 default admin roles.
	//
	// Each new tenant gets a generic role catalog: "Super Admin" (full
	// perms) and "Operational Admin" (minus users:manage / roles:manage /
	// settings:manage). Names are NOT prefixed with the company — the
	// (type, name, company_id) unique index keeps them disambiguated, and
	// the UI flow is "pick company → pick role within that company" rather
	// than browsing a flat global list of role names.
	//
	// The legacy SuperAdminRoleID=1 short-circuit in auth/models.go grants
	// blanket permission regardless of company; for non-id=1 super admins
	// we instead grant explicit permissions so the handler-level
	// company_id filter is the sole isolation boundary.
	//
	// User records (info@rule27design.com, robert@rule27design.com) are
	// left for Alchemy to create via the admin UI after the migration
	// runs — passwords are set there.
	// --------------------------------------------------------------------
	if _, err := tx.Exec(`
		INSERT INTO roles (type, name, permissions, company_id)
		VALUES (
			'user',
			'Super Admin',
			ARRAY[
				'lists:get_all', 'lists:manage_all',
				'subscribers:get_all', 'subscribers:get', 'subscribers:manage', 'subscribers:import', 'subscribers:sql_query',
				'campaigns:get_all', 'campaigns:manage_all', 'campaigns:get', 'campaigns:get_analytics', 'campaigns:manage',
				'tx:send',
				'bounces:get', 'bounces:manage',
				'media:get', 'media:manage',
				'templates:get', 'templates:manage',
				'users:get', 'users:manage',
				'roles:get', 'roles:manage',
				'settings:get', 'settings:manage', 'settings:maintain',
				'segments:get', 'segments:manage',
				'webhooks:get', 'webhooks:manage',
				'drips:get', 'drips:manage',
				'deals:get', 'deals:manage',
				'activities:get', 'activities:manage',
				'webhooks:post_bounce'
			]::TEXT[],
			2
		)
		ON CONFLICT (type, name, company_id) WHERE name IS NOT NULL DO NOTHING;

		INSERT INTO roles (type, name, permissions, company_id)
		VALUES (
			'user',
			'Operational Admin',
			ARRAY[
				'lists:get_all', 'lists:manage_all',
				'subscribers:get_all', 'subscribers:get', 'subscribers:manage', 'subscribers:import',
				'campaigns:get_all', 'campaigns:manage_all', 'campaigns:get', 'campaigns:get_analytics', 'campaigns:manage',
				'tx:send',
				'bounces:get', 'bounces:manage',
				'media:get', 'media:manage',
				'templates:get', 'templates:manage',
				'segments:get', 'segments:manage',
				'webhooks:get',
				'drips:get', 'drips:manage',
				'deals:get', 'deals:manage',
				'activities:get', 'activities:manage'
			]::TEXT[],
			2
		)
		ON CONFLICT (type, name, company_id) WHERE name IS NOT NULL DO NOTHING;
	`); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	lo.Println("v7.17.0 migration complete.")
	lo.Println("  → Companies seeded: Solomon (id=1), Rule27 Design (id=2)")
	lo.Println("  → Rule27 default admin roles created: Super Admin, Operational Admin (under company_id=2)")
	lo.Println("  → NEXT: create info@rule27design.com + robert@rule27design.com via the admin UI,")
	lo.Println("         assigning them to company_id=2 and the appropriate role.")
	lo.Println("  → To enable hard isolation: set `app.enforce_company_isolation = true` in config.toml")
	lo.Println("    after smoke-testing. Until then, handlers run in dual-mode (no filtering).")

	return nil
}
