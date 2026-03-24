package core

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

var dripQuerySortFields = []string{"name", "status", "created_at", "updated_at"}

// QueryDripCampaigns retrieves paginated drip campaigns.
func (c *Core) QueryDripCampaigns(searchStr, orderBy, order string, offset, limit int) (models.DripCampaigns, int, error) {
	queryStr, stmt := makeSearchQuery(searchStr, orderBy, order, c.q.QueryDripCampaigns, dripQuerySortFields)

	var out models.DripCampaigns
	if err := c.db.Select(&out, stmt, 0, nil, queryStr, offset, limit); err != nil {
		c.log.Printf("error fetching drip campaigns: %v", err)
		return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "drip campaigns", "error", pqErrMsg(err)))
	}

	total := 0
	if len(out) > 0 {
		total = out[0].Total
	}

	return out, total, nil
}

// GetDripCampaign retrieves a drip campaign by ID or UUID.
func (c *Core) GetDripCampaign(id int, uuid string) (models.DripCampaign, error) {
	var uu any
	if uuid != "" {
		uu = uuid
	}

	var out models.DripCampaign
	if err := c.q.GetDripCampaign.Get(&out, id, uu); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "drip campaign", "error", pqErrMsg(err)))
	}

	if out.ID == 0 {
		return out, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "drip campaign"))
	}

	// Load steps.
	steps, err := c.GetDripSteps(out.ID)
	if err != nil {
		return out, err
	}
	out.Steps = steps

	return out, nil
}

// CreateDripCampaign creates a new drip campaign.
func (c *Core) CreateDripCampaign(o models.DripCampaign) (models.DripCampaign, error) {
	uu, err := uuid.NewV4()
	if err != nil {
		c.log.Printf("error generating UUID: %v", err)
		return models.DripCampaign{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "drip campaign", "error", err.Error()))
	}

	if o.Status == "" {
		o.Status = models.DripStatusDraft
	}
	if o.TriggerType == "" {
		o.TriggerType = models.DripTriggerSubscription
	}
	if o.TriggerConfig == nil {
		o.TriggerConfig = []byte("{}")
	}

	var id int
	if err := c.q.CreateDripCampaign.Get(&id, uu, o.Name, o.Description, o.Status, o.TriggerType, o.TriggerConfig, o.SegmentID, o.FromEmail); err != nil {
		c.log.Printf("error creating drip campaign: %v", err)
		return models.DripCampaign{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "drip campaign", "error", pqErrMsg(err)))
	}

	return c.GetDripCampaign(id, "")
}

// UpdateDripCampaign updates a drip campaign.
func (c *Core) UpdateDripCampaign(id int, o models.DripCampaign) (models.DripCampaign, error) {
	if o.TriggerConfig == nil {
		o.TriggerConfig = []byte("{}")
	}

	res, err := c.q.UpdateDripCampaign.Exec(id, o.Name, o.Description, o.Status, o.TriggerType, o.TriggerConfig, o.SegmentID, o.FromEmail)
	if err != nil {
		c.log.Printf("error updating drip campaign: %v", err)
		return models.DripCampaign{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "drip campaign", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.DripCampaign{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "drip campaign"))
	}

	return c.GetDripCampaign(id, "")
}

// UpdateDripCampaignStatus updates just the status of a drip campaign.
func (c *Core) UpdateDripCampaignStatus(id int, status string) error {
	_, err := c.q.UpdateDripCampaignStatus.Exec(id, status)
	if err != nil {
		c.log.Printf("error updating drip campaign status: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "drip campaign", "error", pqErrMsg(err)))
	}
	return nil
}

// DeleteDripCampaign deletes a drip campaign.
func (c *Core) DeleteDripCampaign(id int) error {
	res, err := c.q.DeleteDripCampaign.Exec(id)
	if err != nil {
		c.log.Printf("error deleting drip campaign: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "drip campaign", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "drip campaign"))
	}
	return nil
}

// GetDripSteps returns all steps for a drip campaign ordered by sequence.
func (c *Core) GetDripSteps(dripCampaignID int) ([]models.DripStep, error) {
	var out []models.DripStep
	if err := c.q.GetDripSteps.Select(&out, dripCampaignID); err != nil {
		c.log.Printf("error fetching drip steps: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "drip steps", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetDripStep returns a single drip step.
func (c *Core) GetDripStep(id int) (models.DripStep, error) {
	var out models.DripStep
	if err := c.q.GetDripStep.Get(&out, id); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "drip step", "error", pqErrMsg(err)))
	}
	return out, nil
}

// CreateDripStep creates a new drip step.
func (c *Core) CreateDripStep(o models.DripStep) (models.DripStep, error) {
	uu, err := uuid.NewV4()
	if err != nil {
		return models.DripStep{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "drip step", "error", err.Error()))
	}

	if o.Headers == nil {
		o.Headers = []byte("[]")
	}
	if o.SendConditions == nil {
		o.SendConditions = []byte("[]")
	}
	if o.Messenger == "" {
		o.Messenger = "email"
	}
	if o.ContentType == "" {
		o.ContentType = "richtext"
	}

	var id int
	if err := c.q.CreateDripStep.Get(&id, uu, o.DripCampaignID, o.SequenceOrder, o.DelayValue, o.DelayUnit,
		o.Name, o.Subject, o.FromEmail, o.Body, o.AltBody, o.ContentType, o.TemplateID, o.Messenger, o.Headers, o.SendConditions); err != nil {
		c.log.Printf("error creating drip step: %v", err)
		return models.DripStep{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "drip step", "error", pqErrMsg(err)))
	}

	return c.GetDripStep(id)
}

// UpdateDripStep updates a drip step.
func (c *Core) UpdateDripStep(id int, o models.DripStep) (models.DripStep, error) {
	if o.Headers == nil {
		o.Headers = []byte("[]")
	}
	if o.SendConditions == nil {
		o.SendConditions = []byte("[]")
	}

	res, err := c.q.UpdateDripStep.Exec(id, o.SequenceOrder, o.DelayValue, o.DelayUnit,
		o.Name, o.Subject, o.FromEmail, o.Body, o.AltBody, o.ContentType, o.TemplateID, o.Messenger, o.Headers, o.SendConditions)
	if err != nil {
		c.log.Printf("error updating drip step: %v", err)
		return models.DripStep{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "drip step", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.DripStep{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "drip step"))
	}

	return c.GetDripStep(id)
}

// DeleteDripStep deletes a drip step.
func (c *Core) DeleteDripStep(id int) error {
	res, err := c.q.DeleteDripStep.Exec(id)
	if err != nil {
		c.log.Printf("error deleting drip step: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "drip step", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "drip step"))
	}
	return nil
}

// EnrollSubscriberInDrip enrolls a subscriber in a drip campaign starting at the first step.
func (c *Core) EnrollSubscriberInDrip(dripCampaignID, subscriberID int) error {
	steps, err := c.GetDripSteps(dripCampaignID)
	if err != nil {
		return err
	}
	if len(steps) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "drip campaign has no steps")
	}

	firstStep := steps[0]
	nextSendAt := calculateNextSend(firstStep.DelayValue, firstStep.DelayUnit)

	if _, err := c.q.EnrollSubscriberInDrip.Exec(dripCampaignID, subscriberID, firstStep.ID, nextSendAt); err != nil {
		c.log.Printf("error enrolling subscriber %d in drip %d: %v", subscriberID, dripCampaignID, err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "drip enrollment", "error", pqErrMsg(err)))
	}

	return nil
}

// GetPendingDripSends returns drip sends that are ready to be dispatched.
func (c *Core) GetPendingDripSends(limit int) ([]models.PendingDripSend, error) {
	var out []models.PendingDripSend
	if err := c.q.GetPendingDripSends.Select(&out, limit); err != nil {
		c.log.Printf("error fetching pending drip sends: %v", err)
		return nil, err
	}
	return out, nil
}

// AdvanceDripEnrollment moves an enrollment to the next step or marks it complete.
func (c *Core) AdvanceDripEnrollment(enrollmentID int64, currentStepID, dripCampaignID int) error {
	steps, err := c.GetDripSteps(dripCampaignID)
	if err != nil {
		return err
	}

	// Find the index of the current step.
	currentIdx := -1
	for i, s := range steps {
		if s.ID == currentStepID {
			currentIdx = i
			break
		}
	}

	if currentIdx == -1 || currentIdx >= len(steps)-1 {
		// No more steps — mark enrollment as completed.
		_, err := c.q.AdvanceDripEnrollment.Exec(enrollmentID, nil, nil, "completed")
		return err
	}

	// Advance to the next step.
	nextStep := steps[currentIdx+1]
	nextSendAt := calculateNextSend(nextStep.DelayValue, nextStep.DelayUnit)
	_, err = c.q.AdvanceDripEnrollment.Exec(enrollmentID, nextStep.ID, nextSendAt, "active")
	return err
}

// RecordDripSend logs a drip send event.
func (c *Core) RecordDripSend(dripCampaignID, stepID, subscriberID int, status, errMsg string) {
	if _, err := c.q.InsertDripSendLog.Exec(dripCampaignID, stepID, subscriberID, status, errMsg); err != nil {
		c.log.Printf("error recording drip send: %v", err)
	}
}

// GetDripEnrollments returns paginated enrollments for a drip campaign.
func (c *Core) GetDripEnrollments(dripCampaignID, offset, limit int) (models.DripEnrollments, int, error) {
	var out models.DripEnrollments
	if err := c.db.Select(&out, c.q.GetDripEnrollments, dripCampaignID, offset, limit); err != nil {
		c.log.Printf("error fetching drip enrollments: %v", err)
		return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "drip enrollments", "error", pqErrMsg(err)))
	}

	total := 0
	if len(out) > 0 {
		total = out[0].Total
	}
	return out, total, nil
}

// GetActiveDripsByTrigger returns all active drip campaigns with a specific trigger type.
func (c *Core) GetActiveDripsByTrigger(triggerType string) (models.DripCampaigns, error) {
	var out models.DripCampaigns
	if err := c.q.GetActiveDripsByTrigger.Select(&out, triggerType); err != nil {
		c.log.Printf("error fetching drips by trigger: %v", err)
		return nil, err
	}
	return out, nil
}

// calculateNextSend calculates the next send time based on delay value and unit.
func calculateNextSend(delayValue int, delayUnit string) time.Time {
	now := time.Now()
	switch delayUnit {
	case "minutes":
		return now.Add(time.Duration(delayValue) * time.Minute)
	case "hours":
		return now.Add(time.Duration(delayValue) * time.Hour)
	case "days":
		return now.AddDate(0, 0, delayValue)
	case "weeks":
		return now.AddDate(0, 0, delayValue*7)
	default:
		return now.AddDate(0, 0, delayValue) // default to days
	}
}
