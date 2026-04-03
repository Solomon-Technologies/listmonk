package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_12_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.12.0 migration: per-campaign messenger selection ...")

	if _, err := db.Exec(`
		ALTER TABLE warming_campaigns ADD COLUMN IF NOT EXISTS messenger VARCHAR DEFAULT '' NOT NULL;
	`); err != nil {
		return err
	}

	return nil
}
