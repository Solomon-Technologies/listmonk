package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_11_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.11.0 migration: per-sender warming campaigns ...")

	if _, err := db.Exec(`
		ALTER TABLE warming_campaigns ADD COLUMN IF NOT EXISTS sender_id INTEGER REFERENCES warming_senders(id);
	`); err != nil {
		return err
	}

	return nil
}
