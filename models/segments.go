package models

import (
	"encoding/json"

	"github.com/lib/pq"
	null "gopkg.in/volatiletech/null.v6"
)

// Segment statuses.
const (
	SegmentMatchAll = "all"
	SegmentMatchAny = "any"
)

// Segments represents a slice of Segment.
type Segments []Segment

// Segment represents a saved subscriber segment with dynamic conditions.
type Segment struct {
	Base

	UUID            string         `db:"uuid" json:"uuid"`
	Name            string         `db:"name" json:"name"`
	Description     string         `db:"description" json:"description"`
	MatchType       string         `db:"match_type" json:"match_type"`
	Conditions      SegmentConds   `db:"conditions" json:"conditions"`
	SubscriberCount int            `db:"subscriber_count" json:"subscriber_count"`
	Tags            pq.StringArray `db:"tags" json:"tags"`

	// Pseudofield for total count in paginated queries.
	Total int `db:"total" json:"-"`
}

// SegmentConds represents a JSON array of segment conditions.
type SegmentConds []SegmentCondition

// SegmentCondition represents a single filter condition in a segment.
type SegmentCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// Valid operators for segment conditions.
var SegmentOperators = map[string]bool{
	"eq": true, "neq": true, "contains": true, "not_contains": true,
	"gt": true, "lt": true, "gte": true, "lte": true,
	"starts_with": true, "ends_with": true,
	"is_set": true, "is_not_set": true,
	"in_list": true, "not_in_list": true,
}

// Valid fields for segment conditions.
var SegmentFields = map[string]bool{
	"email": true, "name": true, "status": true,
	"created_at": true, "updated_at": true,
	"lists": true, "score": true,
}

// Scan implements the sql.Scanner interface for SegmentConds.
func (s *SegmentConds) Scan(src any) error {
	if src == nil {
		*s = make(SegmentConds, 0)
		return nil
	}
	if b, ok := src.([]byte); ok {
		return json.Unmarshal(b, s)
	}
	return nil
}

// Value implements the driver.Valuer interface for SegmentConds.
func (s SegmentConds) Value() (any, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

// GetIDs returns the list of segment IDs.
func (segs Segments) GetIDs() []int {
	IDs := make([]int, len(segs))
	for i, s := range segs {
		IDs[i] = s.ID
	}
	return IDs
}

// Dummy null types for optional fields.
var _ null.String
