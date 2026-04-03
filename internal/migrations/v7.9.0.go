package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_9_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.9.0 migration: warming campaigns ...")

	if _, err := db.Exec(`
		-- Warming campaigns (individual warming campaigns per brand/domain)
		CREATE TABLE IF NOT EXISTS warming_campaigns (
			id                 SERIAL PRIMARY KEY,
			name               TEXT NOT NULL,
			brand              TEXT NOT NULL DEFAULT '',
			sender_domains     TEXT[] NOT NULL DEFAULT '{}',
			status             TEXT NOT NULL DEFAULT 'draft',
			sends_per_run      INTEGER NOT NULL DEFAULT 3,
			runs_per_day       INTEGER NOT NULL DEFAULT 4,
			schedule_times     TEXT[] NOT NULL DEFAULT '{"10:00","14:00","18:00","21:00"}',
			random_delay_min_s INTEGER NOT NULL DEFAULT 30,
			random_delay_max_s INTEGER NOT NULL DEFAULT 120,
			created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Add campaign_id to send log for per-campaign tracking
		ALTER TABLE warming_send_log ADD COLUMN IF NOT EXISTS campaign_id INTEGER REFERENCES warming_campaigns(id) ON DELETE SET NULL;
		CREATE INDEX IF NOT EXISTS idx_warming_send_log_campaign ON warming_send_log(campaign_id);
	`); err != nil {
		return err
	}

	// Seed two warming campaigns matching existing sender domains.
	db.Exec(`INSERT INTO warming_campaigns (name, brand, sender_domains, status, sends_per_run, runs_per_day, schedule_times)
		VALUES ('SolomonTech Warming', 'Solomon Technology', '{"solomontech.co"}', 'active', 3, 4, '{"10:00","14:00","18:00","21:00"}')
		ON CONFLICT DO NOTHING`)
	db.Exec(`INSERT INTO warming_campaigns (name, brand, sender_domains, status, sends_per_run, runs_per_day, schedule_times)
		VALUES ('AniltX Warming', 'AniltX', '{"aniltx.biz","aniltx.pro"}', 'active', 3, 4, '{"10:00","14:00","18:00","21:00"}')
		ON CONFLICT DO NOTHING`)

	return nil
}
