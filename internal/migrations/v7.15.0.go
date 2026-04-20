package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V7_15_0: evergreen campaigns.
//
// Adds `is_evergreen` to the campaigns table. When true, the manager keeps
// the campaign in `running` status even after subscribers are exhausted,
// and a separate scanner re-queues it so newly-added list subscribers get
// the campaign on their next eligibility pass.
//
// Also adds a partial uniqueness index on campaign_send_log (campaign_id,
// subscriber_id) so the NOT EXISTS dedup the evergreen subscriber query
// relies on is indexed. Partial — only applies to non-null subscriber_id,
// which is everything except the historical rows preserved by v7.14.0.
func V7_15_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.15.0 migration: evergreen campaigns ...")

	if _, err := db.Exec(`
		ALTER TABLE campaigns
		ADD COLUMN IF NOT EXISTS is_evergreen BOOLEAN NOT NULL DEFAULT false;
	`); err != nil {
		return err
	}
	if _, err := db.Exec(`
		COMMENT ON COLUMN campaigns.is_evergreen IS
		  'Evergreen campaigns stay in running status after the initial send drain and re-queue for new list subscribers on a scheduled rescan.';
	`); err != nil {
		return err
	}

	// Send-log dedup index — the evergreen subscriber-fetch query joins
	// against this. Partial so we only index rows where subscriber_id is
	// still set (v7.14.0 switched FK to ON DELETE SET NULL).
	if _, err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_csl_campaign_subscriber_unique
		ON campaign_send_log (campaign_id, subscriber_id)
		WHERE subscriber_id IS NOT NULL;
	`); err != nil {
		return err
	}

	return nil
}
