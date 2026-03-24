package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_6_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.6.0 migration: CRM settings ...")

	if _, err := db.Exec(`
		INSERT INTO settings (key, value, updated_at)
		VALUES
			('crm.deal_stages', '["Lead", "Qualified", "Proposal", "Negotiation", "Closed Won", "Closed Lost"]', NOW()),
			('crm.currencies', '["USD", "EUR", "GBP"]', NOW()),
			('crm.activity_types', '["note", "call", "meeting", "email", "task"]', NOW()),
			('crm.default_deal_stage', '"Lead"', NOW()),
			('crm.default_currency', '"USD"', NOW())
		ON CONFLICT (key) DO NOTHING;
	`); err != nil {
		return err
	}

	return nil
}
