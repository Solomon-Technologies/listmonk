package models

import (
	"time"

	"github.com/jmoiron/sqlx/types"
	null "gopkg.in/volatiletech/null.v6"
)

// Drip campaign statuses.
const (
	DripStatusDraft    = "draft"
	DripStatusActive   = "active"
	DripStatusPaused   = "paused"
	DripStatusArchived = "archived"
)

// Drip trigger types.
const (
	DripTriggerSubscription = "subscription"
	DripTriggerSegmentEntry = "segment_entry"
	DripTriggerTagAdded     = "tag_added"
	DripTriggerDateField    = "date_field"
	DripTriggerManual       = "manual"
)

// DripCampaigns represents a slice of DripCampaign.
type DripCampaigns []DripCampaign

// DripCampaign represents a drip/automation email campaign.
type DripCampaign struct {
	Base

	UUID           string         `db:"uuid" json:"uuid"`
	Name           string         `db:"name" json:"name"`
	Description    string         `db:"description" json:"description"`
	Status         string         `db:"status" json:"status"`
	TriggerType    string         `db:"trigger_type" json:"trigger_type"`
	TriggerConfig  types.JSONText `db:"trigger_config" json:"trigger_config"`
	SegmentID      null.Int       `db:"segment_id" json:"segment_id"`
	FromEmail      string         `db:"from_email" json:"from_email"`
	TotalEntered   int            `db:"total_entered" json:"total_entered"`
	TotalCompleted int            `db:"total_completed" json:"total_completed"`
	TotalExited    int            `db:"total_exited" json:"total_exited"`

	// Loaded separately.
	Steps []DripStep `db:"-" json:"steps,omitempty"`

	// Pseudofield for total count in paginated queries.
	Total int `db:"total" json:"-"`
}

// DripStep represents a single step in a drip campaign.
type DripStep struct {
	Base

	UUID           string         `db:"uuid" json:"uuid"`
	DripCampaignID int            `db:"drip_campaign_id" json:"drip_campaign_id"`
	SequenceOrder  int            `db:"sequence_order" json:"sequence_order"`
	DelayValue     int            `db:"delay_value" json:"delay_value"`
	DelayUnit      string         `db:"delay_unit" json:"delay_unit"`
	Name           string         `db:"name" json:"name"`
	Subject        string         `db:"subject" json:"subject"`
	FromEmail      string         `db:"from_email" json:"from_email"`
	Body           string         `db:"body" json:"body"`
	AltBody        string         `db:"alt_body" json:"alt_body"`
	ContentType    string         `db:"content_type" json:"content_type"`
	TemplateID     null.Int       `db:"template_id" json:"template_id"`
	Messenger      string         `db:"messenger" json:"messenger"`
	Headers        types.JSONText `db:"headers" json:"headers"`
	SendConditions types.JSONText `db:"send_conditions" json:"send_conditions"`
	Sent           int            `db:"sent" json:"sent"`
	Opened         int            `db:"opened" json:"opened"`
	Clicked        int            `db:"clicked" json:"clicked"`
}

// DripEnrollment represents a subscriber's enrollment in a drip campaign.
type DripEnrollment struct {
	ID              int64     `db:"id" json:"id"`
	DripCampaignID  int       `db:"drip_campaign_id" json:"drip_campaign_id"`
	SubscriberID    int       `db:"subscriber_id" json:"subscriber_id"`
	Status          string    `db:"status" json:"status"`
	CurrentStepID   null.Int  `db:"current_step_id" json:"current_step_id"`
	NextSendAt      null.Time `db:"next_send_at" json:"next_send_at"`
	EnteredAt       time.Time `db:"entered_at" json:"entered_at"`
	CompletedAt     null.Time `db:"completed_at" json:"completed_at"`
	SubscriberEmail string    `db:"subscriber_email" json:"subscriber_email,omitempty"`
	SubscriberName  string    `db:"subscriber_name" json:"subscriber_name,omitempty"`

	// Pseudofield for total count in paginated queries.
	Total int `db:"total" json:"-"`
}

// DripEnrollments represents a slice of DripEnrollment.
type DripEnrollments []DripEnrollment

// PendingDripSend represents a drip email that is ready to be sent.
type PendingDripSend struct {
	EnrollmentID      int64          `db:"enrollment_id"`
	DripCampaignID    int            `db:"drip_campaign_id"`
	SubscriberID      int            `db:"subscriber_id"`
	CurrentStepID     int            `db:"current_step_id"`
	SubscriberEmail   string         `db:"subscriber_email"`
	SubscriberName    string         `db:"subscriber_name"`
	SubscriberUUID    string         `db:"subscriber_uuid"`
	SubscriberAttribs types.JSONText `db:"subscriber_attribs"`
	SubscriberStatus  string         `db:"subscriber_status"`
	Subject           string         `db:"subject"`
	Body              string         `db:"body"`
	AltBody           string         `db:"alt_body"`
	StepFromEmail     string         `db:"step_from_email"`
	ContentType       string         `db:"content_type"`
	TemplateID        null.Int       `db:"template_id"`
	Messenger         string         `db:"messenger"`
	Headers           types.JSONText `db:"headers"`
	CampaignFromEmail string         `db:"campaign_from_email"`
	CampaignName      string         `db:"campaign_name"`
}

// DripSendLog represents a drip send log entry.
type DripSendLog struct {
	ID              int64     `db:"id" json:"id"`
	DripCampaignID  int       `db:"drip_campaign_id" json:"drip_campaign_id"`
	DripStepID      int       `db:"drip_step_id" json:"drip_step_id"`
	SubscriberID    int       `db:"subscriber_id" json:"subscriber_id"`
	Status          string    `db:"status" json:"status"`
	ErrorMessage    string    `db:"error_message" json:"error_message"`
	SentAt          time.Time `db:"sent_at" json:"sent_at"`
}

// Valid drip statuses.
var DripStatuses = map[string]bool{
	DripStatusDraft: true, DripStatusActive: true,
	DripStatusPaused: true, DripStatusArchived: true,
}

// Valid drip trigger types.
var DripTriggerTypes = map[string]bool{
	DripTriggerSubscription: true, DripTriggerSegmentEntry: true,
	DripTriggerTagAdded: true, DripTriggerDateField: true, DripTriggerManual: true,
}
