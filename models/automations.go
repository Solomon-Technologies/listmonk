package models

import (
	"time"

	"github.com/jmoiron/sqlx/types"
	null "gopkg.in/volatiletech/null.v6"
)

// Automation statuses.
const (
	AutomationStatusDraft    = "draft"
	AutomationStatusActive   = "active"
	AutomationStatusPaused   = "paused"
	AutomationStatusArchived = "archived"
)

// Automation node types.
const (
	NodeTypeTrigger       = "trigger"
	NodeTypeActionEmail   = "action_email"
	NodeTypeActionTag     = "action_tag"
	NodeTypeActionList    = "action_list"
	NodeTypeActionWebhook = "action_webhook"
	NodeTypeCondition     = "condition"
	NodeTypeDelay         = "delay"
	NodeTypeWaitFor       = "wait_for"
	NodeTypeSplit         = "split"
)

// Automations represents a slice of Automation.
type Automations []Automation

// Automation represents a visual automation workflow.
type Automation struct {
	Base

	UUID           string         `db:"uuid" json:"uuid"`
	Name           string         `db:"name" json:"name"`
	Description    string         `db:"description" json:"description"`
	Status         string         `db:"status" json:"status"`
	Canvas         types.JSONText `db:"canvas" json:"canvas"`
	TotalEntered   int            `db:"total_entered" json:"total_entered"`
	TotalCompleted int            `db:"total_completed" json:"total_completed"`

	// Loaded separately.
	Nodes []AutomationNode `db:"-" json:"nodes,omitempty"`
	Edges []AutomationEdge `db:"-" json:"edges,omitempty"`

	// Pseudofield for total count in paginated queries.
	Total int `db:"total" json:"-"`
}

// AutomationNode represents a single node in an automation workflow.
type AutomationNode struct {
	ID           int            `db:"id" json:"id"`
	UUID         string         `db:"uuid" json:"uuid"`
	AutomationID int            `db:"automation_id" json:"automation_id"`
	NodeType     string         `db:"node_type" json:"node_type"`
	Config       types.JSONText `db:"config" json:"config"`
	PositionX    int            `db:"position_x" json:"position_x"`
	PositionY    int            `db:"position_y" json:"position_y"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

// AutomationEdge represents a connection between two nodes.
type AutomationEdge struct {
	ID           int    `db:"id" json:"id"`
	AutomationID int    `db:"automation_id" json:"automation_id"`
	FromNodeID   int    `db:"from_node_id" json:"from_node_id"`
	ToNodeID     int    `db:"to_node_id" json:"to_node_id"`
	Label        string `db:"label" json:"label"`
}

// AutomationEnrollment represents a subscriber in an automation.
type AutomationEnrollment struct {
	ID            int64     `db:"id" json:"id"`
	AutomationID  int       `db:"automation_id" json:"automation_id"`
	SubscriberID  int       `db:"subscriber_id" json:"subscriber_id"`
	CurrentNodeID null.Int  `db:"current_node_id" json:"current_node_id"`
	Status        string    `db:"status" json:"status"`
	WaitUntil     null.Time `db:"wait_until" json:"wait_until"`
	EnteredAt     time.Time `db:"entered_at" json:"entered_at"`
	CompletedAt   null.Time `db:"completed_at" json:"completed_at"`

	// Joined fields.
	SubscriberEmail   string         `db:"subscriber_email" json:"subscriber_email,omitempty"`
	SubscriberName    string         `db:"subscriber_name" json:"subscriber_name,omitempty"`
	SubscriberUUID    string         `db:"subscriber_uuid" json:"subscriber_uuid,omitempty"`
	SubscriberAttribs types.JSONText `db:"subscriber_attribs" json:"subscriber_attribs,omitempty"`

	// Pseudofield.
	EnrollmentID int64 `db:"enrollment_id" json:"-"`
}

// Valid node types.
var AutomationNodeTypes = map[string]bool{
	NodeTypeTrigger: true, NodeTypeActionEmail: true, NodeTypeActionTag: true,
	NodeTypeActionList: true, NodeTypeActionWebhook: true, NodeTypeCondition: true,
	NodeTypeDelay: true, NodeTypeWaitFor: true, NodeTypeSplit: true,
}
