package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_7_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.7.0 migration: drip production enhancements ...")

	if _, err := db.Exec(`
		ALTER TABLE drip_campaigns ADD COLUMN IF NOT EXISTS max_send_per_day INTEGER NOT NULL DEFAULT 0;
	`); err != nil {
		return err
	}

	return nil
}
