package models

import (
	"time"

	null "gopkg.in/volatiletech/null.v6"
)

// A/B test statuses.
const (
	ABTestStatusDraft    = "draft"
	ABTestStatusRunning  = "running"
	ABTestStatusFinished = "finished"
)

// A/B test types.
const (
	ABTestTypeSubject  = "subject"
	ABTestTypeContent  = "content"
	ABTestTypeFrom     = "from_email"
	ABTestTypeSendTime = "send_time"
)

// A/B winner metrics.
const (
	ABMetricOpenRate  = "open_rate"
	ABMetricClickRate = "click_rate"
)

// ABTest represents an A/B test on a campaign.
type ABTest struct {
	Base

	UUID             string    `db:"uuid" json:"uuid"`
	CampaignID       int       `db:"campaign_id" json:"campaign_id"`
	TestType         string    `db:"test_type" json:"test_type"`
	Status           string    `db:"status" json:"status"`
	TestPercentage   int       `db:"test_percentage" json:"test_percentage"`
	WinnerMetric     string    `db:"winner_metric" json:"winner_metric"`
	WinnerWaitHours  int       `db:"winner_wait_hours" json:"winner_wait_hours"`
	WinningVariantID null.Int  `db:"winning_variant_id" json:"winning_variant_id"`
	StartedAt        null.Time `db:"started_at" json:"started_at"`
	FinishedAt       null.Time `db:"finished_at" json:"finished_at"`

	// Loaded separately.
	Variants []ABTestVariant `db:"-" json:"variants,omitempty"`
}

// ABTestVariant represents a variant in an A/B test.
type ABTestVariant struct {
	ID         int       `db:"id" json:"id"`
	ABTestID   int       `db:"ab_test_id" json:"ab_test_id"`
	Label      string    `db:"label" json:"label"`
	Subject    string    `db:"subject" json:"subject"`
	Body       string    `db:"body" json:"body"`
	FromEmail  string    `db:"from_email" json:"from_email"`
	Sent       int       `db:"sent" json:"sent"`
	Opened     int       `db:"opened" json:"opened"`
	Clicked    int       `db:"clicked" json:"clicked"`
	Bounced    int       `db:"bounced" json:"bounced"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`

	// Computed fields.
	OpenRate  float64 `db:"-" json:"open_rate"`
	ClickRate float64 `db:"-" json:"click_rate"`
}

// ABTestAssignment represents a subscriber-to-variant mapping.
type ABTestAssignment struct {
	ID           int64 `db:"id" json:"id"`
	ABTestID     int   `db:"ab_test_id" json:"ab_test_id"`
	VariantID    int   `db:"variant_id" json:"variant_id"`
	SubscriberID int   `db:"subscriber_id" json:"subscriber_id"`
}

// Valid A/B test types.
var ABTestTypes = map[string]bool{
	ABTestTypeSubject: true, ABTestTypeContent: true,
	ABTestTypeFrom: true, ABTestTypeSendTime: true,
}

// Valid A/B winner metrics.
var ABWinnerMetrics = map[string]bool{
	ABMetricOpenRate: true, ABMetricClickRate: true,
}
