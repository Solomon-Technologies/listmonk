package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_0_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.0.0 migration: segments + webhooks ...")

	// Phase 1A: Segments
	if _, err := db.Exec(`
		DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'segment_match') THEN
				CREATE TYPE segment_match AS ENUM ('all', 'any');
			END IF;
		END $$;

		CREATE TABLE IF NOT EXISTS segments (
			id              SERIAL PRIMARY KEY,
			uuid            uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
			name            TEXT NOT NULL,
			description     TEXT NOT NULL DEFAULT '',
			match_type      segment_match NOT NULL DEFAULT 'all',
			conditions      JSONB NOT NULL DEFAULT '[]',
			subscriber_count INT NOT NULL DEFAULT 0,
			tags            VARCHAR(100)[],
			created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_segments_name ON segments(name);
	`); err != nil {
		return err
	}

	// Phase 1B: Webhooks
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS webhooks (
			id              SERIAL PRIMARY KEY,
			uuid            uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
			name            TEXT NOT NULL,
			url             TEXT NOT NULL,
			secret          TEXT NOT NULL DEFAULT '',
			enabled         BOOLEAN NOT NULL DEFAULT true,
			events          TEXT[] NOT NULL DEFAULT '{}',
			max_retries     INT NOT NULL DEFAULT 3,
			timeout_seconds INT NOT NULL DEFAULT 10,
			total_sent      INT NOT NULL DEFAULT 0,
			total_failed    INT NOT NULL DEFAULT 0,
			last_error      TEXT,
			last_sent_at    TIMESTAMP WITH TIME ZONE,
			created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_webhooks_enabled ON webhooks(enabled);

		CREATE TABLE IF NOT EXISTS webhook_log (
			id              BIGSERIAL PRIMARY KEY,
			webhook_id      INTEGER NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
			event           TEXT NOT NULL,
			payload         JSONB NOT NULL DEFAULT '{}',
			response_code   INT,
			response_body   TEXT,
			error           TEXT,
			attempt         INT NOT NULL DEFAULT 1,
			created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_webhook_log_webhook ON webhook_log(webhook_id);
		CREATE INDEX IF NOT EXISTS idx_webhook_log_date ON webhook_log(created_at);
	`); err != nil {
		return err
	}

	// Add segment and webhook permissions to existing roles.
	if _, err := db.Exec(`
		UPDATE user_roles SET permissions = permissions ||
			ARRAY['segments:get', 'segments:manage', 'webhooks:get', 'webhooks:manage']
		WHERE permissions @> ARRAY['settings:manage']
		AND NOT (permissions @> ARRAY['segments:get'])
	`); err != nil {
		// Non-fatal: permissions might not exist yet or column format differs.
		lo.Printf("note: could not auto-add segment/webhook permissions to admin roles: %v", err)
	}

	return nil
}
