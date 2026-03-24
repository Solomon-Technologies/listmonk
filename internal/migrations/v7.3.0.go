package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_3_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.3.0 migration: visual automation builder ...")

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS automations (
			id               SERIAL PRIMARY KEY,
			uuid             uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
			name             TEXT NOT NULL,
			description      TEXT NOT NULL DEFAULT '',
			status           TEXT NOT NULL DEFAULT 'draft',
			canvas           JSONB NOT NULL DEFAULT '{"nodes":[],"edges":[]}',
			total_entered    INT NOT NULL DEFAULT 0,
			total_completed  INT NOT NULL DEFAULT 0,
			created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_automations_status ON automations(status);

		CREATE TABLE IF NOT EXISTS automation_nodes (
			id               SERIAL PRIMARY KEY,
			uuid             uuid NOT NULL UNIQUE DEFAULT gen_random_uuid(),
			automation_id    INT NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
			node_type        TEXT NOT NULL,
			config           JSONB NOT NULL DEFAULT '{}',
			position_x       INT NOT NULL DEFAULT 0,
			position_y       INT NOT NULL DEFAULT 0,
			created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_auto_nodes_automation ON automation_nodes(automation_id);

		CREATE TABLE IF NOT EXISTS automation_edges (
			id               SERIAL PRIMARY KEY,
			automation_id    INT NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
			from_node_id     INT NOT NULL REFERENCES automation_nodes(id) ON DELETE CASCADE,
			to_node_id       INT NOT NULL REFERENCES automation_nodes(id) ON DELETE CASCADE,
			label            TEXT NOT NULL DEFAULT '',
			UNIQUE(automation_id, from_node_id, to_node_id)
		);

		CREATE TABLE IF NOT EXISTS automation_enrollments (
			id               BIGSERIAL PRIMARY KEY,
			automation_id    INT NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
			subscriber_id    INT NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
			current_node_id  INT REFERENCES automation_nodes(id) ON DELETE SET NULL,
			status           TEXT NOT NULL DEFAULT 'active',
			wait_until       TIMESTAMP WITH TIME ZONE,
			entered_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			completed_at     TIMESTAMP WITH TIME ZONE,
			UNIQUE(automation_id, subscriber_id)
		);
		CREATE INDEX IF NOT EXISTS idx_auto_enroll_wait ON automation_enrollments(wait_until) WHERE status = 'active';
	`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		UPDATE user_roles SET permissions = permissions ||
			ARRAY['automations:get', 'automations:manage']
		WHERE permissions @> ARRAY['settings:manage']
		AND NOT (permissions @> ARRAY['automations:get'])
	`); err != nil {
		lo.Printf("note: could not auto-add automation permissions: %v", err)
	}

	return nil
}
