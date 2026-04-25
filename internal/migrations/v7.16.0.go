package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V7_16_0: per-campaign warming recipient selection.
//
// Adds `recipient_ids` to warming_campaigns. When non-empty, the warming
// processor sends only to addresses whose id is in this array. When empty
// (default), the campaign sends to every active warming_addresses row,
// preserving prior behavior.
func V7_16_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.16.0 migration: per-campaign warming recipient selection ...")

	if _, err := db.Exec(`
		ALTER TABLE warming_campaigns
		ADD COLUMN IF NOT EXISTS recipient_ids INT[] NOT NULL DEFAULT '{}';
	`); err != nil {
		return err
	}

	return nil
}
