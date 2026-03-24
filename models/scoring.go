package models

import (
	"time"

	"github.com/jmoiron/sqlx/types"
	null "gopkg.in/volatiletech/null.v6"
)

// Scoring event types.
const (
	ScoreEventEmailOpened   = "email.opened"
	ScoreEventEmailClicked  = "email.clicked"
	ScoreEventEmailBounced  = "email.bounced"
	ScoreEventListSubscribed   = "list.subscribed"
	ScoreEventListUnsubscribed = "list.unsubscribed"
	ScoreEventInactivity30  = "inactivity.30days"
)

// ScoringRule represents a scoring rule configuration.
type ScoringRule struct {
	ID         int            `db:"id" json:"id"`
	Name       string         `db:"name" json:"name"`
	Enabled    bool           `db:"enabled" json:"enabled"`
	EventType  string         `db:"event_type" json:"event_type"`
	ScoreValue int            `db:"score_value" json:"score_value"`
	Conditions types.JSONText `db:"conditions" json:"conditions"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at" json:"updated_at"`
}

// ScoringRules represents a slice of ScoringRule.
type ScoringRules []ScoringRule

// ScoreLog represents a score change event.
type ScoreLog struct {
	ID           int64          `db:"id" json:"id"`
	SubscriberID int            `db:"subscriber_id" json:"subscriber_id"`
	RuleID       null.Int       `db:"rule_id" json:"rule_id"`
	EventType    string         `db:"event_type" json:"event_type"`
	ScoreChange  int            `db:"score_change" json:"score_change"`
	ScoreAfter   int            `db:"score_after" json:"score_after"`
	Meta         types.JSONText `db:"meta" json:"meta"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

// ScoreLogs represents a slice of ScoreLog.
type ScoreLogs []ScoreLog

// Valid scoring events.
var ScoringEvents = map[string]bool{
	ScoreEventEmailOpened: true, ScoreEventEmailClicked: true,
	ScoreEventEmailBounced: true, ScoreEventListSubscribed: true,
	ScoreEventListUnsubscribed: true, ScoreEventInactivity30: true,
}
