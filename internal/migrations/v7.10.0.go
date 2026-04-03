package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_10_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.10.0 migration: warming progressive ramp + hourly cap + business hours ...")

	if _, err := db.Exec(`
		ALTER TABLE warming_campaigns ADD COLUMN IF NOT EXISTS warmup_start_date DATE;
		ALTER TABLE warming_campaigns ADD COLUMN IF NOT EXISTS daily_limits JSONB NOT NULL DEFAULT '[]';
		ALTER TABLE warming_campaigns ADD COLUMN IF NOT EXISTS hourly_cap INTEGER NOT NULL DEFAULT 0;
		ALTER TABLE warming_campaigns ADD COLUMN IF NOT EXISTS business_hours_only BOOLEAN NOT NULL DEFAULT false;
	`); err != nil {
		return err
	}

	return nil
}
