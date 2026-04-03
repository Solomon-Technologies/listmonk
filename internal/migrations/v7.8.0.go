package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V7_8_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("running Solomon v7.8.0 migration: email warming feature ...")

	if _, err := db.Exec(`
		-- Warming addresses (recipients for warming emails)
		CREATE TABLE IF NOT EXISTS warming_addresses (
			id         SERIAL PRIMARY KEY,
			email      TEXT NOT NULL UNIQUE,
			name       TEXT NOT NULL DEFAULT '',
			is_active  BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Warming senders (from addresses for warming emails)
		CREATE TABLE IF NOT EXISTS warming_senders (
			id          SERIAL PRIMARY KEY,
			email       TEXT NOT NULL UNIQUE,
			name        TEXT NOT NULL DEFAULT '',
			brand       TEXT NOT NULL DEFAULT '',
			brand_url   TEXT NOT NULL DEFAULT '',
			brand_color TEXT NOT NULL DEFAULT '#F2C94C',
			is_active   BOOLEAN NOT NULL DEFAULT true,
			created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Warming templates (conversational text templates)
		CREATE TABLE IF NOT EXISTS warming_templates (
			id         SERIAL PRIMARY KEY,
			subject    TEXT NOT NULL,
			body       TEXT NOT NULL,
			is_active  BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Warming config (singleton row with schedule and limits)
		CREATE TABLE IF NOT EXISTS warming_config (
			id                 INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
			sends_per_run      INTEGER NOT NULL DEFAULT 3,
			runs_per_day       INTEGER NOT NULL DEFAULT 4,
			schedule_times     TEXT[] NOT NULL DEFAULT '{"10:00","14:00","18:00","21:00"}',
			random_delay_min_s INTEGER NOT NULL DEFAULT 30,
			random_delay_max_s INTEGER NOT NULL DEFAULT 120,
			is_active          BOOLEAN NOT NULL DEFAULT false,
			updated_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		INSERT INTO warming_config (id) VALUES (1) ON CONFLICT DO NOTHING;

		-- Warming send log (audit trail)
		CREATE TABLE IF NOT EXISTS warming_send_log (
			id              BIGSERIAL PRIMARY KEY,
			sender_email    TEXT NOT NULL,
			recipient_email TEXT NOT NULL,
			template_id     INTEGER REFERENCES warming_templates(id) ON DELETE SET NULL,
			subject         TEXT NOT NULL,
			status          TEXT NOT NULL DEFAULT 'sent',
			error_message   TEXT NOT NULL DEFAULT '',
			sent_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_warming_send_log_sent_at ON warming_send_log(sent_at);
	`); err != nil {
		return err
	}

	// Seed default warming addresses.
	addresses := []struct{ email, name string }{
		{"r.alch3my@gmail.com", "R"},
		{"thisisreitaka@gmail.com", "Reitaka"},
		{"myfife20@gmail.com", "Fife"},
		{"blasianps@gmail.com", "B"},
		{"myfifecars@gmail.com", "Fife Cars"},
	}
	for _, a := range addresses {
		db.Exec(`INSERT INTO warming_addresses (email, name) VALUES ($1, $2) ON CONFLICT DO NOTHING`, a.email, a.name)
	}

	// Seed default warming senders.
	senders := []struct{ email, name, brand, brandURL, brandColor string }{
		{"solomon.c@solomontech.co", "Solomon C", "Solomon Technology", "https://www.solomon.technology", "#0ea5e9"},
		{"Alch3my@solomontech.co", "Alchemy", "Solomon Technology", "https://www.solomon.technology", "#0ea5e9"},
		{"mytaneia.f@solomontech.co", "Mytaneia F", "Solomon Technology", "https://www.solomon.technology", "#0ea5e9"},
		{"sales@aniltx.biz", "AnilTX Sales", "AniltX", "https://app.aniltx.com", "#F2C94C"},
		{"partners@aniltx.biz", "AnilTX Partners", "AniltX", "https://app.aniltx.com", "#F2C94C"},
		{"no-reply@aniltx.pro", "AniltX", "AniltX", "https://app.aniltx.com", "#F2C94C"},
	}
	for _, s := range senders {
		db.Exec(`INSERT INTO warming_senders (email, name, brand, brand_url, brand_color) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
			s.email, s.name, s.brand, s.brandURL, s.brandColor)
	}

	// Seed default warming templates.
	templates := []struct{ subject, body string }{
		{"Quick sync — {{date}}", "Hey {{name}}, just checking in. Let me know if you need anything from my end this week. Talk soon."},
		{"Following up on our conversation", "Wanted to follow up from our last chat. I put together some notes — happy to walk through them whenever works for you."},
		{"Thought of you", "Hey {{name}}, saw something that reminded me of our discussion. Would love to reconnect when you have a minute. Hope all is well."},
		{"Quick question", "Hey — quick question for you. Do you have 5 minutes this week? Nothing urgent, just wanted to run something by you."},
		{"Touching base", "Just touching base. Things are moving on our end and I wanted to keep you in the loop. Let me know a good time to chat."},
		{"Re: update", "Quick update from our side — we've been making good progress on a few things. Would love to share when you have a sec."},
		{"Hey — one more thing", "One more thing I forgot to mention last time. Nothing major, but thought you'd find it interesting. Let me know if you want the details."},
		{"Checking in", "Hey {{name}}, been a bit — wanted to check in and see how things are going on your end. Let's catch up soon."},
		{"Got a sec?", "Hey, got a quick sec? Had an idea I wanted to bounce off you. No rush — whenever you're free."},
		{"Loop back", "Circling back on something from earlier. Let me know when you have a minute and I'll fill you in."},
		{"Saw this and thought of you", "Hey {{name}}, came across something relevant to what we discussed. Would love to share — let me know when you have a sec."},
		{"Before I forget", "Hey — wanted to send this before it slips my mind. Quick thing I think you'd find useful. Let me know if you want more details."},
		{"Real quick", "Hey {{name}}, real quick — are you around this week? Got something interesting to run by you. No rush at all."},
		{"Heads up", "Quick heads up — we're working on something I think you'll want to know about. I'll share more soon, but wanted to plant the seed."},
		{"One thing I wanted to mention", "Hey, one thing I meant to bring up earlier. Nothing time-sensitive, just thought it might be relevant to you."},
	}
	for _, t := range templates {
		db.Exec(`INSERT INTO warming_templates (subject, body) VALUES ($1, $2) ON CONFLICT DO NOTHING`, t.subject, t.body)
	}

	return nil
}
