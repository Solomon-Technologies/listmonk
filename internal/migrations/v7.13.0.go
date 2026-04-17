package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V7_13_0 creates campaign_send_log — a per-recipient send record written
// by the manager immediately after the messenger Push(). Powers the UI
// "Send Log" tab + GET /api/campaigns/:id/send-log endpoint so admins can
// see exactly who got which campaign when (and what failed), independent
// of open/click tracking.
func V7_13_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.13.0 migration: campaign_send_log ...")

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS campaign_send_log (
			id               BIGSERIAL PRIMARY KEY,
			campaign_id      INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
			subscriber_id    INTEGER NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
			subscriber_email TEXT NOT NULL,
			sent_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			messenger        TEXT NOT NULL DEFAULT '',
			status           TEXT NOT NULL DEFAULT 'sent',
			error_message    TEXT
		);
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_csl_campaign
			ON campaign_send_log (campaign_id, sent_at DESC);
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_csl_subscriber
			ON campaign_send_log (subscriber_id, sent_at DESC);
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_csl_sent_at
			ON campaign_send_log (sent_at DESC);
	`); err != nil {
		return err
	}

	return nil
}
