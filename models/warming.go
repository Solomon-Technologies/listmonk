package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
	null "gopkg.in/volatiletech/null.v6"
)

// WarmingAddress represents a recipient email for warming sends.
type WarmingAddress struct {
	Base
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// WarmingSender represents a sender email for warming sends.
type WarmingSender struct {
	Base
	Email      string    `db:"email" json:"email"`
	Name       string    `db:"name" json:"name"`
	Brand      string    `db:"brand" json:"brand"`
	BrandURL   string    `db:"brand_url" json:"brand_url"`
	BrandColor string    `db:"brand_color" json:"brand_color"`
	IsActive   bool      `db:"is_active" json:"is_active"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// WarmingTemplate represents a conversational warming email template.
type WarmingTemplate struct {
	Base
	Subject   string    `db:"subject" json:"subject"`
	Body      string    `db:"body" json:"body"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// WarmingConfig represents the singleton warming scheduler config.
type WarmingConfig struct {
	ID             int            `db:"id" json:"id"`
	SendsPerRun    int            `db:"sends_per_run" json:"sends_per_run"`
	RunsPerDay     int            `db:"runs_per_day" json:"runs_per_day"`
	ScheduleTimes  pq.StringArray `db:"schedule_times" json:"schedule_times"`
	RandomDelayMin int            `db:"random_delay_min_s" json:"random_delay_min_s"`
	RandomDelayMax int            `db:"random_delay_max_s" json:"random_delay_max_s"`
	IsActive       bool           `db:"is_active" json:"is_active"`
	UpdatedAt      null.Time      `db:"updated_at" json:"updated_at"`
}

// WarmingCampaign represents an individual warming campaign per brand/domain.
type WarmingCampaign struct {
	Base
	Name           string         `db:"name" json:"name"`
	Brand          string         `db:"brand" json:"brand"`
	SenderDomains  pq.StringArray `db:"sender_domains" json:"sender_domains"`
	Status         string         `db:"status" json:"status"`
	SendsPerRun    int            `db:"sends_per_run" json:"sends_per_run"`
	RunsPerDay     int            `db:"runs_per_day" json:"runs_per_day"`
	ScheduleTimes  pq.StringArray `db:"schedule_times" json:"schedule_times"`
	RandomDelayMin    int              `db:"random_delay_min_s" json:"random_delay_min_s"`
	RandomDelayMax    int              `db:"random_delay_max_s" json:"random_delay_max_s"`
	WarmupStartDate   null.Time        `db:"warmup_start_date" json:"warmup_start_date"`
	DailyLimits       json.RawMessage  `db:"daily_limits" json:"daily_limits"`
	HourlyCap         int              `db:"hourly_cap" json:"hourly_cap"`
	BusinessHoursOnly bool             `db:"business_hours_only" json:"business_hours_only"`
	SenderID          null.Int         `db:"sender_id" json:"sender_id"`
	Messenger         string           `db:"messenger" json:"messenger"`
	RecipientIDs      pq.Int64Array    `db:"recipient_ids" json:"recipient_ids"`
	CreatedAt         time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time        `db:"updated_at" json:"updated_at"`
}

// DailyLimit represents a single day's send limit in the progressive ramp schedule.
type DailyLimit struct {
	Day int `json:"day"`
	Max int `json:"max"`
}

// WarmingSendLog represents a warming send log entry.
type WarmingSendLog struct {
	ID             int64     `db:"id" json:"id"`
	CampaignID     null.Int  `db:"campaign_id" json:"campaign_id"`
	SenderEmail    string    `db:"sender_email" json:"sender_email"`
	RecipientEmail string    `db:"recipient_email" json:"recipient_email"`
	TemplateID     null.Int  `db:"template_id" json:"template_id"`
	Subject        string    `db:"subject" json:"subject"`
	Status         string    `db:"status" json:"status"`
	ErrorMessage   string    `db:"error_message" json:"error_message"`
	SentAt         time.Time `db:"sent_at" json:"sent_at"`
	CampaignName   string    `db:"campaign_name" json:"campaign_name,omitempty"`
}

// WarmingStats represents warming send statistics.
type WarmingStats struct {
	SentToday int  `json:"sent_today"`
	IsActive  bool `json:"is_active"`
}

// WarmingCampaignStats represents per-campaign send statistics.
type WarmingCampaignStats struct {
	SentToday   int `db:"sent_today" json:"sent_today"`
	ErrorsToday int `db:"errors_today" json:"errors_today"`
	TotalSent   int `db:"total_sent" json:"total_sent"`
	TotalErrors int `db:"total_errors" json:"total_errors"`
}
