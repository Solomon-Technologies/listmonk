package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_1_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.1.0 migration: drip campaigns ...")

	if _, err := db.Exec(`
		DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'drip_status') THEN
				CREATE TYPE drip_status AS ENUM ('draft', 'active', 'paused', 'archived');
			END IF;
		END $$;

		DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'drip_trigger_type') THEN
				CREATE TYPE drip_trigger_type AS ENUM ('subscription', 'segment_entry', 'tag_added', 'date_field', 'manual');
			END IF;
		END $$;

		CREATE TABLE IF NOT EXISTS drip_campaigns (
			id               SERIAL PRIMARY KEY,
			uuid             uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
			name             TEXT NOT NULL,
			description      TEXT NOT NULL DEFAULT '',
			status           drip_status NOT NULL DEFAULT 'draft',
			trigger_type     drip_trigger_type NOT NULL DEFAULT 'subscription',
			trigger_config   JSONB NOT NULL DEFAULT '{}',
			segment_id       INT NULL REFERENCES segments(id) ON DELETE SET NULL,
			from_email       TEXT NOT NULL DEFAULT '',
			total_entered    INT NOT NULL DEFAULT 0,
			total_completed  INT NOT NULL DEFAULT 0,
			total_exited     INT NOT NULL DEFAULT 0,
			created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_drip_campaigns_status ON drip_campaigns(status);

		CREATE TABLE IF NOT EXISTS drip_steps (
			id                  SERIAL PRIMARY KEY,
			uuid                uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
			drip_campaign_id    INT NOT NULL REFERENCES drip_campaigns(id) ON DELETE CASCADE,
			sequence_order      INT NOT NULL DEFAULT 0,
			delay_value         INT NOT NULL DEFAULT 0,
			delay_unit          TEXT NOT NULL DEFAULT 'days',
			name                TEXT NOT NULL DEFAULT '',
			subject             TEXT NOT NULL DEFAULT '',
			from_email          TEXT NOT NULL DEFAULT '',
			body                TEXT NOT NULL DEFAULT '',
			alt_body            TEXT NOT NULL DEFAULT '',
			content_type        TEXT NOT NULL DEFAULT 'richtext',
			template_id         INT NULL REFERENCES templates(id) ON DELETE SET NULL,
			messenger           TEXT NOT NULL DEFAULT 'email',
			headers             JSONB NOT NULL DEFAULT '[]',
			send_conditions     JSONB NOT NULL DEFAULT '[]',
			sent                INT NOT NULL DEFAULT 0,
			opened              INT NOT NULL DEFAULT 0,
			clicked             INT NOT NULL DEFAULT 0,
			created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_drip_steps_campaign ON drip_steps(drip_campaign_id, sequence_order);

		CREATE TABLE IF NOT EXISTS drip_enrollments (
			id                  BIGSERIAL PRIMARY KEY,
			drip_campaign_id    INT NOT NULL REFERENCES drip_campaigns(id) ON DELETE CASCADE,
			subscriber_id       INT NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
			status              TEXT NOT NULL DEFAULT 'active',
			current_step_id     INT NULL REFERENCES drip_steps(id) ON DELETE SET NULL,
			next_send_at        TIMESTAMP WITH TIME ZONE,
			entered_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			completed_at        TIMESTAMP WITH TIME ZONE,
			UNIQUE(drip_campaign_id, subscriber_id)
		);
		CREATE INDEX IF NOT EXISTS idx_drip_enroll_next ON drip_enrollments(next_send_at) WHERE status = 'active';
		CREATE INDEX IF NOT EXISTS idx_drip_enroll_sub ON drip_enrollments(subscriber_id);

		CREATE TABLE IF NOT EXISTS drip_send_log (
			id                  BIGSERIAL PRIMARY KEY,
			drip_campaign_id    INT REFERENCES drip_campaigns(id) ON DELETE CASCADE,
			drip_step_id        INT REFERENCES drip_steps(id) ON DELETE CASCADE,
			subscriber_id       INT REFERENCES subscribers(id) ON DELETE CASCADE,
			status              TEXT NOT NULL DEFAULT 'sent',
			error_message       TEXT,
			sent_at             TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_drip_send_log_campaign ON drip_send_log(drip_campaign_id);
	`); err != nil {
		return err
	}

	// Add drip permissions to admin roles.
	if _, err := db.Exec(`
		UPDATE user_roles SET permissions = permissions ||
			ARRAY['drips:get', 'drips:manage']
		WHERE permissions @> ARRAY['settings:manage']
		AND NOT (permissions @> ARRAY['drips:get'])
	`); err != nil {
		lo.Printf("note: could not auto-add drip permissions to admin roles: %v", err)
	}

	return nil
}
