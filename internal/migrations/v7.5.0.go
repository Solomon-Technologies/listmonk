package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_5_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.5.0 migration: enhanced analytics + templates + CRM ...")

	// Enhanced analytics: add tracking columns.
	if _, err := db.Exec(`
		ALTER TABLE campaign_views ADD COLUMN IF NOT EXISTS user_agent TEXT;
		ALTER TABLE campaign_views ADD COLUMN IF NOT EXISTS device_type TEXT;
		ALTER TABLE campaign_views ADD COLUMN IF NOT EXISTS email_client TEXT;
		ALTER TABLE campaign_views ADD COLUMN IF NOT EXISTS country TEXT;

		ALTER TABLE link_clicks ADD COLUMN IF NOT EXISTS user_agent TEXT;
		ALTER TABLE link_clicks ADD COLUMN IF NOT EXISTS device_type TEXT;
		ALTER TABLE link_clicks ADD COLUMN IF NOT EXISTS email_client TEXT;
		ALTER TABLE link_clicks ADD COLUMN IF NOT EXISTS country TEXT;

		-- Conversion tracking.
		CREATE TABLE IF NOT EXISTS conversion_goals (
			id          SERIAL PRIMARY KEY,
			name        TEXT NOT NULL,
			url_pattern TEXT NOT NULL,
			enabled     BOOLEAN NOT NULL DEFAULT true,
			created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS conversions (
			id            BIGSERIAL PRIMARY KEY,
			goal_id       INT REFERENCES conversion_goals(id) ON DELETE SET NULL,
			campaign_id   INT REFERENCES campaigns(id) ON DELETE SET NULL,
			subscriber_id INT REFERENCES subscribers(id) ON DELETE SET NULL,
			url           TEXT NOT NULL,
			value         DECIMAL(10,2) DEFAULT 0,
			created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_conversions_campaign ON conversions(campaign_id);
		CREATE INDEX IF NOT EXISTS idx_conversions_goal ON conversions(goal_id);
	`); err != nil {
		return err
	}

	// Template improvements.
	if _, err := db.Exec(`
		ALTER TABLE templates ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '';
		ALTER TABLE templates ADD COLUMN IF NOT EXISTS tags VARCHAR(100)[];
		ALTER TABLE templates ADD COLUMN IF NOT EXISTS thumbnail TEXT;
		ALTER TABLE templates ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
	`); err != nil {
		return err
	}

	// CRM: subscriber extensions + deals + activities.
	if _, err := db.Exec(`
		ALTER TABLE subscribers ADD COLUMN IF NOT EXISTS company TEXT NOT NULL DEFAULT '';
		ALTER TABLE subscribers ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '';
		ALTER TABLE subscribers ADD COLUMN IF NOT EXISTS lifecycle_stage TEXT NOT NULL DEFAULT 'subscriber';

		CREATE TABLE IF NOT EXISTS deals (
			id              SERIAL PRIMARY KEY,
			uuid            uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
			subscriber_id   INT NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
			name            TEXT NOT NULL,
			value           DECIMAL(12,2) NOT NULL DEFAULT 0,
			currency        TEXT NOT NULL DEFAULT 'USD',
			status          TEXT NOT NULL DEFAULT 'open',
			stage           TEXT NOT NULL DEFAULT '',
			expected_close  DATE,
			closed_at       TIMESTAMP WITH TIME ZONE,
			notes           TEXT NOT NULL DEFAULT '',
			attribs         JSONB NOT NULL DEFAULT '{}',
			created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_deals_subscriber ON deals(subscriber_id);
		CREATE INDEX IF NOT EXISTS idx_deals_status ON deals(status);

		CREATE TABLE IF NOT EXISTS contact_activities (
			id              BIGSERIAL PRIMARY KEY,
			subscriber_id   INT NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
			activity_type   TEXT NOT NULL,
			description     TEXT NOT NULL DEFAULT '',
			meta            JSONB NOT NULL DEFAULT '{}',
			created_by      INT REFERENCES users(id) ON DELETE SET NULL,
			created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_contact_activities_sub ON contact_activities(subscriber_id);
		CREATE INDEX IF NOT EXISTS idx_contact_activities_date ON contact_activities(created_at);
	`); err != nil {
		return err
	}

	// Add CRM permissions.
	if _, err := db.Exec(`
		UPDATE user_roles SET permissions = permissions ||
			ARRAY['deals:get', 'deals:manage', 'activities:get', 'activities:manage']
		WHERE permissions @> ARRAY['settings:manage']
		AND NOT (permissions @> ARRAY['deals:get'])
	`); err != nil {
		lo.Printf("note: could not auto-add CRM permissions: %v", err)
	}

	return nil
}
