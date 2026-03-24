package models

import (
	"time"

	"github.com/jmoiron/sqlx/types"
	null "gopkg.in/volatiletech/null.v6"
)

// Deal statuses.
const (
	DealStatusOpen   = "open"
	DealStatusWon    = "won"
	DealStatusLost   = "lost"
)

// Deals represents a slice of Deal.
type Deals []Deal

// Deal represents a CRM deal/opportunity.
type Deal struct {
	Base

	UUID          string         `db:"uuid" json:"uuid"`
	SubscriberID  int            `db:"subscriber_id" json:"subscriber_id"`
	Name          string         `db:"name" json:"name"`
	Value         float64        `db:"value" json:"value"`
	Currency      string         `db:"currency" json:"currency"`
	Status        string         `db:"status" json:"status"`
	Stage         string         `db:"stage" json:"stage"`
	ExpectedClose null.Time      `db:"expected_close" json:"expected_close"`
	ClosedAt      null.Time      `db:"closed_at" json:"closed_at"`
	Notes         string         `db:"notes" json:"notes"`
	Attribs       types.JSONText `db:"attribs" json:"attribs"`

	// Pseudofield for total count in paginated queries.
	Total int `db:"total" json:"-"`
}

// ContactActivity represents a CRM activity log entry.
type ContactActivity struct {
	ID            int64          `db:"id" json:"id"`
	SubscriberID  int            `db:"subscriber_id" json:"subscriber_id"`
	ActivityType  string         `db:"activity_type" json:"activity_type"`
	Description   string         `db:"description" json:"description"`
	Meta          types.JSONText `db:"meta" json:"meta"`
	CreatedBy     null.Int       `db:"created_by" json:"created_by"`
	CreatedByName string         `db:"created_by_name" json:"created_by_name,omitempty"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`

	// Pseudofield for total count in paginated queries.
	Total int `db:"total" json:"-"`
}

// ContactActivities represents a slice of ContactActivity.
type ContactActivities []ContactActivity

// DealPipelineEntry represents an aggregated pipeline view.
type DealPipelineEntry struct {
	Status     string  `db:"status" json:"status"`
	Stage      string  `db:"stage" json:"stage"`
	Count      int     `db:"count" json:"count"`
	TotalValue float64 `db:"total_value" json:"total_value"`
}

// Valid deal statuses.
var DealStatuses = map[string]bool{
	DealStatusOpen: true, DealStatusWon: true, DealStatusLost: true,
}

// Valid activity types.
var ActivityTypes = map[string]bool{
	"note": true, "email": true, "call": true, "meeting": true, "task": true,
}
