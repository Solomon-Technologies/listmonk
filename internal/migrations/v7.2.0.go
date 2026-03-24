package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_2_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.2.0 migration: A/B testing ...")

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ab_tests (
			id                SERIAL PRIMARY KEY,
			uuid              uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
			campaign_id       INT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
			test_type         TEXT NOT NULL DEFAULT 'subject',
			status            TEXT NOT NULL DEFAULT 'draft',
			test_percentage   INT NOT NULL DEFAULT 20 CHECK (test_percentage BETWEEN 5 AND 50),
			winner_metric     TEXT NOT NULL DEFAULT 'open_rate',
			winner_wait_hours INT NOT NULL DEFAULT 4,
			winning_variant_id INT,
			started_at        TIMESTAMP WITH TIME ZONE,
			finished_at       TIMESTAMP WITH TIME ZONE,
			created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_ab_tests_campaign ON ab_tests(campaign_id);
		CREATE INDEX IF NOT EXISTS idx_ab_tests_status ON ab_tests(status);

		CREATE TABLE IF NOT EXISTS ab_test_variants (
			id          SERIAL PRIMARY KEY,
			ab_test_id  INT NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			label       TEXT NOT NULL DEFAULT 'A',
			subject     TEXT NOT NULL DEFAULT '',
			body        TEXT NOT NULL DEFAULT '',
			from_email  TEXT NOT NULL DEFAULT '',
			sent        INT NOT NULL DEFAULT 0,
			opened      INT NOT NULL DEFAULT 0,
			clicked     INT NOT NULL DEFAULT 0,
			bounced     INT NOT NULL DEFAULT 0,
			created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_ab_variants_test ON ab_test_variants(ab_test_id);

		CREATE TABLE IF NOT EXISTS ab_test_assignments (
			id            BIGSERIAL PRIMARY KEY,
			ab_test_id    INT NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
			variant_id    INT NOT NULL REFERENCES ab_test_variants(id) ON DELETE CASCADE,
			subscriber_id INT NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
			UNIQUE(ab_test_id, subscriber_id)
		);
		CREATE INDEX IF NOT EXISTS idx_ab_assign_variant ON ab_test_assignments(variant_id);
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		UPDATE user_roles SET permissions = permissions ||
			ARRAY['ab_tests:get', 'ab_tests:manage']
		WHERE permissions @> ARRAY['settings:manage']
		AND NOT (permissions @> ARRAY['ab_tests:get'])
	`); err != nil {
		lo.Printf("note: could not auto-add A/B test permissions: %v", err)
	}

	return nil
}
