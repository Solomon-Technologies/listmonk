package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_4_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.4.0 migration: contact scoring ...")

	if _, err := db.Exec(`
		ALTER TABLE subscribers ADD COLUMN IF NOT EXISTS score INT NOT NULL DEFAULT 0;
		CREATE INDEX IF NOT EXISTS idx_subs_score ON subscribers(score);

		CREATE TABLE IF NOT EXISTS scoring_rules (
			id           SERIAL PRIMARY KEY,
			name         TEXT NOT NULL,
			enabled      BOOLEAN NOT NULL DEFAULT true,
			event_type   TEXT NOT NULL,
			score_value  INT NOT NULL DEFAULT 0,
			conditions   JSONB NOT NULL DEFAULT '{}',
			created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS score_log (
			id              BIGSERIAL PRIMARY KEY,
			subscriber_id   INT NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
			rule_id         INT REFERENCES scoring_rules(id) ON DELETE SET NULL,
			event_type      TEXT NOT NULL,
			score_change    INT NOT NULL,
			score_after     INT NOT NULL,
			meta            JSONB NOT NULL DEFAULT '{}',
			created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_score_log_sub ON score_log(subscriber_id);
		CREATE INDEX IF NOT EXISTS idx_score_log_date ON score_log(created_at);
	`); err != nil {
		return err
	}

	// Insert default scoring rules.
	if _, err := db.Exec(`
		INSERT INTO scoring_rules (name, event_type, score_value) VALUES
			('Email Opened', 'email.opened', 5),
			('Link Clicked', 'email.clicked', 10),
			('Email Bounced', 'email.bounced', -15),
			('List Subscribed', 'list.subscribed', 20),
			('List Unsubscribed', 'list.unsubscribed', -25),
			('30 Day Inactivity', 'inactivity.30days', -10)
		ON CONFLICT DO NOTHING
	`); err != nil {
		lo.Printf("note: could not insert default scoring rules: %v", err)
	}

	if _, err := db.Exec(`
		UPDATE user_roles SET permissions = permissions ||
			ARRAY['scoring:get', 'scoring:manage']
		WHERE permissions @> ARRAY['settings:manage']
		AND NOT (permissions @> ARRAY['scoring:get'])
	`); err != nil {
		lo.Printf("note: could not auto-add scoring permissions: %v", err)
	}

	return nil
}
