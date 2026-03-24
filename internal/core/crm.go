package core

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// QueryDeals retrieves paginated deals.
func (c *Core) QueryDeals(subscriberID int, status string, offset, limit int) (models.Deals, int, error) {
	var out models.Deals
	if err := c.db.Select(&out, c.q.QueryDeals, subscriberID, status, offset, limit); err != nil {
		c.log.Printf("error fetching deals: %v", err)
		return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "deals", "error", pqErrMsg(err)))
	}

	total := 0
	if len(out) > 0 {
		total = out[0].Total
	}

	return out, total, nil
}

// GetDeal retrieves a deal by ID or UUID.
func (c *Core) GetDeal(id int, uuStr string) (models.Deal, error) {
	var uu any
	if uuStr != "" {
		uu = uuStr
	}

	var out models.Deal
	if err := c.q.GetDeal.Get(&out, id, uu); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "deal", "error", pqErrMsg(err)))
	}

	if out.ID == 0 {
		return out, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "deal"))
	}

	return out, nil
}

// CreateDeal creates a new deal.
func (c *Core) CreateDeal(o models.Deal) (models.Deal, error) {
	uu, err := uuid.NewV4()
	if err != nil {
		return models.Deal{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "deal", "error", err.Error()))
	}

	if o.Currency == "" {
		o.Currency = "USD"
	}
	if o.Status == "" {
		o.Status = models.DealStatusOpen
	}
	if o.Attribs == nil {
		o.Attribs = []byte("{}")
	}

	var id int
	if err := c.q.CreateDeal.Get(&id, uu, o.SubscriberID, o.Name, o.Value, o.Currency, o.Status, o.Stage, o.ExpectedClose, o.Notes, o.Attribs); err != nil {
		c.log.Printf("error creating deal: %v", err)
		return models.Deal{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "deal", "error", pqErrMsg(err)))
	}

	return c.GetDeal(id, "")
}

// UpdateDeal updates a deal.
func (c *Core) UpdateDeal(id int, o models.Deal) (models.Deal, error) {
	if o.Attribs == nil {
		o.Attribs = []byte("{}")
	}

	res, err := c.q.UpdateDeal.Exec(id, o.Name, o.Value, o.Currency, o.Status, o.Stage, o.ExpectedClose, o.ClosedAt, o.Notes, o.Attribs)
	if err != nil {
		c.log.Printf("error updating deal: %v", err)
		return models.Deal{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "deal", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.Deal{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "deal"))
	}

	return c.GetDeal(id, "")
}

// DeleteDeal deletes a deal.
func (c *Core) DeleteDeal(id int) error {
	res, err := c.q.DeleteDeal.Exec(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "deal", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "deal"))
	}
	return nil
}

// GetDealPipeline returns the deal pipeline summary.
func (c *Core) GetDealPipeline() ([]models.DealPipelineEntry, error) {
	var out []models.DealPipelineEntry
	if err := c.q.GetDealPipeline.Select(&out); err != nil {
		c.log.Printf("error fetching deal pipeline: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "deal pipeline", "error", pqErrMsg(err)))
	}
	return out, nil
}

// CreateContactActivity creates a new activity log entry.
func (c *Core) CreateContactActivity(o models.ContactActivity) (models.ContactActivity, error) {
	if o.Meta == nil {
		o.Meta = []byte("{}")
	}

	var id int64
	if err := c.q.CreateActivity.Get(&id, o.SubscriberID, o.ActivityType, o.Description, o.Meta, o.CreatedBy); err != nil {
		c.log.Printf("error creating activity: %v", err)
		return models.ContactActivity{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "activity", "error", pqErrMsg(err)))
	}

	o.ID = id
	return o, nil
}

// GetSubscriberActivities returns paginated activities for a subscriber.
func (c *Core) GetSubscriberActivities(subscriberID, offset, limit int) (models.ContactActivities, int, error) {
	var out models.ContactActivities
	if err := c.db.Select(&out, c.q.GetSubscriberActivities, subscriberID, offset, limit); err != nil {
		c.log.Printf("error fetching activities: %v", err)
		return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "activities", "error", pqErrMsg(err)))
	}

	total := 0
	if len(out) > 0 {
		total = out[0].Total
	}

	return out, total, nil
}

// DeleteActivity deletes a contact activity.
func (c *Core) DeleteActivity(id int64) error {
	_, err := c.q.DeleteActivity.Exec(id)
	return err
}
