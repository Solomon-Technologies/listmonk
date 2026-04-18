package models

import (
	"time"

	"github.com/lib/pq"
	null "gopkg.in/volatiletech/null.v6"
)

// Webhooks represents a slice of Webhook.
type Webhooks []Webhook

// Webhook represents a webhook endpoint configuration.
type Webhook struct {
	Base

	UUID           string         `db:"uuid" json:"uuid"`
	Name           string         `db:"name" json:"name"`
	URL            string         `db:"url" json:"url"`
	Secret         string         `db:"secret" json:"secret,omitempty"`
	Enabled        bool           `db:"enabled" json:"enabled"`
	Events         pq.StringArray `db:"events" json:"events"`
	MaxRetries     int            `db:"max_retries" json:"max_retries"`
	TimeoutSeconds int            `db:"timeout_seconds" json:"timeout_seconds"`
	TotalSent      int            `db:"total_sent" json:"total_sent"`
	TotalFailed    int            `db:"total_failed" json:"total_failed"`
	LastError      null.String    `db:"last_error" json:"last_error"`
	LastSentAt     null.Time      `db:"last_sent_at" json:"last_sent_at"`

	// Pseudofield for total count in paginated queries.
	Total int `db:"total" json:"-"`
}

// WebhookLog represents a single webhook delivery log entry.
type WebhookLog struct {
	ID           int       `db:"id" json:"id"`
	WebhookID    int       `db:"webhook_id" json:"webhook_id"`
	Event        string    `db:"event" json:"event"`
	Payload      JSON      `db:"payload" json:"payload"`
	ResponseCode null.Int  `db:"response_code" json:"response_code"`
	ResponseBody null.String `db:"response_body" json:"response_body"`
	Error        null.String `db:"error" json:"error"`
	Attempt      int       `db:"attempt" json:"attempt"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`

	// Pseudofield for total count in paginated queries.
	Total int `db:"total" json:"-"`
}

// WebhookLogs represents a slice of WebhookLog.
type WebhookLogs []WebhookLog

// WebhookEvents lists all supported webhook event types.
var WebhookEvents = map[string]bool{
	"subscriber.created":      true,
	"subscriber.updated":      true,
	"subscriber.optin":        true,
	"subscriber.unsubscribed": true,
	"campaign.started":        true,
	"campaign.finished":       true,
	"campaign.view":           true,
	"campaign.click":          true,
	"campaign.send":           true,
	"bounce.received":         true,
	"drip.enrolled":           true,
	"drip.step_sent":          true,
	"drip.completed":          true,
}

// GetIDs returns the list of webhook IDs.
func (whs Webhooks) GetIDs() []int {
	IDs := make([]int, len(whs))
	for i, w := range whs {
		IDs[i] = w.ID
	}
	return IDs
}
