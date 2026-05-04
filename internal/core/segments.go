package core

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

var segmentQuerySortFields = []string{"name", "created_at", "updated_at", "subscriber_count"}

// QuerySegments retrieves paginated segments optionally filtering by search.
// companyID=0 disables tenant filtering.
func (c *Core) QuerySegments(searchStr string, tags []string, orderBy, order string, offset, limit, companyID int) (models.Segments, int, error) {
	queryStr, stmt := makeSearchQuery(searchStr, orderBy, order, c.q.QuerySegments, segmentQuerySortFields)

	if tags == nil {
		tags = []string{}
	}

	var out models.Segments
	if err := c.db.Select(&out, stmt, 0, nil, queryStr, pq.StringArray(tags), offset, limit, companyID); err != nil {
		c.log.Printf("error fetching segments: %v", err)
		return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "segments", "error", pqErrMsg(err)))
	}

	for i := range out {
		if out[i].Tags == nil {
			out[i].Tags = []string{}
		}
		if out[i].Conditions == nil {
			out[i].Conditions = models.SegmentConds{}
		}
	}

	total := 0
	if len(out) > 0 {
		total = out[0].Total
	}

	return out, total, nil
}

// GetSegment retrieves a segment by ID or UUID.
// companyID=0 disables tenant filtering.
func (c *Core) GetSegment(id int, uuid string, companyID int) (models.Segment, error) {
	var uu any
	if uuid != "" {
		uu = uuid
	}

	var out models.Segment
	if err := c.q.GetSegment.Get(&out, id, uu, companyID); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "segment", "error", pqErrMsg(err)))
	}

	if out.ID == 0 {
		return out, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "segment"))
	}

	if out.Tags == nil {
		out.Tags = []string{}
	}
	if out.Conditions == nil {
		out.Conditions = models.SegmentConds{}
	}

	return out, nil
}

// CreateSegment creates a new segment. companyID stamps tenant.
func (c *Core) CreateSegment(o models.Segment, companyID int) (models.Segment, error) {
	uu, err := uuid.NewV4()
	if err != nil {
		c.log.Printf("error generating UUID: %v", err)
		return models.Segment{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "segment", "error", err.Error()))
	}

	if o.Conditions == nil {
		o.Conditions = models.SegmentConds{}
	}
	if o.Tags == nil {
		o.Tags = []string{}
	}

	var id int
	if err := c.q.CreateSegment.Get(&id, uu, o.Name, o.Description, o.MatchType, o.Conditions, pq.StringArray(o.Tags), companyID); err != nil {
		c.log.Printf("error creating segment: %v", err)
		return models.Segment{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "segment", "error", pqErrMsg(err)))
	}

	return c.GetSegment(id, "", 0)
}

// UpdateSegment updates a segment.
func (c *Core) UpdateSegment(id int, o models.Segment) (models.Segment, error) {
	if o.Conditions == nil {
		o.Conditions = models.SegmentConds{}
	}
	if o.Tags == nil {
		o.Tags = []string{}
	}

	res, err := c.q.UpdateSegment.Exec(id, o.Name, o.Description, o.MatchType, o.Conditions, pq.StringArray(o.Tags))
	if err != nil {
		c.log.Printf("error updating segment: %v", err)
		return models.Segment{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "segment", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.Segment{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "segment"))
	}

	return c.GetSegment(id, "", 0)
}

// DeleteSegment deletes a segment.
func (c *Core) DeleteSegment(id int) error {
	res, err := c.q.DeleteSegment.Exec(id)
	if err != nil {
		c.log.Printf("error deleting segment: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "segment", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "segment"))
	}

	return nil
}

// GetSegmentSubscriberCount dynamically counts subscribers matching a segment's conditions.
func (c *Core) GetSegmentSubscriberCount(seg models.Segment) (int, error) {
	where, err := BuildSegmentWhere(seg.Conditions, seg.MatchType)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid segment conditions: %v", err))
	}

	var count int
	q := fmt.Sprintf(`SELECT COUNT(*) FROM subscribers WHERE status != 'blocklisted' AND (%s)`, where)
	if err := c.db.Get(&count, q); err != nil {
		c.log.Printf("error counting segment subscribers: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "segment count", "error", pqErrMsg(err)))
	}

	// Update the cached count.
	c.q.UpdateSegmentCount.Exec(seg.ID, count)

	return count, nil
}

// BuildSegmentWhere builds a SQL WHERE clause from segment conditions.
// It uses parameterized values where possible and validates all field names.
func BuildSegmentWhere(conditions models.SegmentConds, matchType string) (string, error) {
	if len(conditions) == 0 {
		return "TRUE", nil
	}

	joiner := " AND "
	if matchType == models.SegmentMatchAny {
		joiner = " OR "
	}

	var clauses []string
	for _, cond := range conditions {
		clause, err := buildConditionClause(cond)
		if err != nil {
			return "", err
		}
		clauses = append(clauses, clause)
	}

	return strings.Join(clauses, joiner), nil
}

// buildConditionClause converts a single SegmentCondition to a SQL clause.
func buildConditionClause(cond models.SegmentCondition) (string, error) {
	field := cond.Field
	op := cond.Operator

	if !models.SegmentOperators[op] {
		return "", fmt.Errorf("invalid operator: %s", op)
	}

	// Determine the SQL column expression.
	var col string
	if strings.HasPrefix(field, "attribs.") {
		// JSONB attribute access: attribs.company -> attribs->>'company'
		key := strings.TrimPrefix(field, "attribs.")
		key = sanitizeIdentifier(key)
		col = fmt.Sprintf("attribs->>'%s'", key)
	} else if field == "lists" {
		// List membership check uses a subquery.
		return buildListCondition(cond)
	} else if models.SegmentFields[field] {
		col = "subscribers." + field
	} else {
		return "", fmt.Errorf("invalid field: %s", field)
	}

	// Build the comparison.
	val := sanitizeValue(cond.Value)

	switch op {
	case "eq":
		return fmt.Sprintf("%s = '%s'", col, val), nil
	case "neq":
		return fmt.Sprintf("%s != '%s'", col, val), nil
	case "contains":
		return fmt.Sprintf("%s ILIKE '%%%s%%'", col, val), nil
	case "not_contains":
		return fmt.Sprintf("%s NOT ILIKE '%%%s%%'", col, val), nil
	case "gt":
		return fmt.Sprintf("%s > '%s'", col, val), nil
	case "lt":
		return fmt.Sprintf("%s < '%s'", col, val), nil
	case "gte":
		return fmt.Sprintf("%s >= '%s'", col, val), nil
	case "lte":
		return fmt.Sprintf("%s <= '%s'", col, val), nil
	case "starts_with":
		return fmt.Sprintf("%s ILIKE '%s%%'", col, val), nil
	case "ends_with":
		return fmt.Sprintf("%s ILIKE '%%%s'", col, val), nil
	case "is_set":
		return fmt.Sprintf("(%s IS NOT NULL AND %s != '')", col, col), nil
	case "is_not_set":
		return fmt.Sprintf("(%s IS NULL OR %s = '')", col, col), nil
	default:
		return "", fmt.Errorf("unsupported operator: %s", op)
	}
}

// buildListCondition builds a SQL clause for list membership conditions.
func buildListCondition(cond models.SegmentCondition) (string, error) {
	val := sanitizeValue(cond.Value)
	switch cond.Operator {
	case "in_list":
		return fmt.Sprintf("subscribers.id IN (SELECT subscriber_id FROM subscriber_lists WHERE list_id = %s AND status = 'confirmed')", val), nil
	case "not_in_list":
		return fmt.Sprintf("subscribers.id NOT IN (SELECT subscriber_id FROM subscriber_lists WHERE list_id = %s AND status = 'confirmed')", val), nil
	default:
		return "", fmt.Errorf("invalid operator for lists field: %s", cond.Operator)
	}
}

// sanitizeIdentifier removes anything that's not alphanumeric or underscore.
func sanitizeIdentifier(s string) string {
	var out strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			out.WriteRune(c)
		}
	}
	return out.String()
}

// sanitizeValue escapes single quotes for safe SQL string interpolation.
func sanitizeValue(v any) string {
	s := fmt.Sprintf("%v", v)
	return strings.ReplaceAll(s, "'", "''")
}
