package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V7_14_0: make campaign_send_log robust across subscriber deletions.
//
// The v7.13.0 migration used ON DELETE CASCADE on subscriber_id, which
// meant deleting a subscriber destroyed their historical send records —
// breaking audits for subs pruned after a campaign finished. Changing
// the FK to ON DELETE SET NULL preserves the row + the denormalized
// subscriber_email/messenger/status/sent_at fields.
func V7_14_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.14.0 migration: campaign_send_log retention on subscriber delete ...")

	// Make subscriber_id nullable, swap FK to SET NULL.
	if _, err := db.Exec(`
		ALTER TABLE campaign_send_log
		ALTER COLUMN subscriber_id DROP NOT NULL;
	`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		ALTER TABLE campaign_send_log
		DROP CONSTRAINT IF EXISTS campaign_send_log_subscriber_id_fkey;
	`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		ALTER TABLE campaign_send_log
		ADD CONSTRAINT campaign_send_log_subscriber_id_fkey
		FOREIGN KEY (subscriber_id) REFERENCES subscribers(id) ON DELETE SET NULL;
	`); err != nil {
		return err
	}

	return nil
}
