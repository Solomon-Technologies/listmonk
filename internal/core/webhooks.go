package core

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

var webhookQuerySortFields = []string{"name", "created_at", "updated_at"}

// QueryWebhooks retrieves paginated webhooks optionally filtering by search.
func (c *Core) QueryWebhooks(searchStr, orderBy, order string, offset, limit int) (models.Webhooks, int, error) {
	queryStr, stmt := makeSearchQuery(searchStr, orderBy, order, c.q.QueryWebhooks, webhookQuerySortFields)

	var out models.Webhooks
	if err := c.db.Select(&out, stmt, 0, nil, queryStr, offset, limit); err != nil {
		c.log.Printf("error fetching webhooks: %v", err)
		return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "webhooks", "error", pqErrMsg(err)))
	}

	for i := range out {
		if out[i].Events == nil {
			out[i].Events = pq.StringArray{}
		}
	}

	total := 0
	if len(out) > 0 {
		total = out[0].Total
	}

	return out, total, nil
}

// GetWebhook retrieves a webhook by ID or UUID.
func (c *Core) GetWebhook(id int, uuid string) (models.Webhook, error) {
	var uu any
	if uuid != "" {
		uu = uuid
	}

	var out models.Webhook
	if err := c.q.GetWebhook.Get(&out, id, uu); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "webhook", "error", pqErrMsg(err)))
	}

	if out.ID == 0 {
		return out, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "webhook"))
	}

	if out.Events == nil {
		out.Events = pq.StringArray{}
	}

	return out, nil
}

// CreateWebhook creates a new webhook.
func (c *Core) CreateWebhook(o models.Webhook) (models.Webhook, error) {
	uu, err := uuid.NewV4()
	if err != nil {
		c.log.Printf("error generating UUID: %v", err)
		return models.Webhook{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "webhook", "error", err.Error()))
	}

	if o.Events == nil {
		o.Events = pq.StringArray{}
	}

	var id int
	if err := c.q.CreateWebhook.Get(&id, uu, o.Name, o.URL, o.Secret, o.Enabled, pq.StringArray(o.Events), o.MaxRetries, o.TimeoutSeconds); err != nil {
		c.log.Printf("error creating webhook: %v", err)
		return models.Webhook{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "webhook", "error", pqErrMsg(err)))
	}

	return c.GetWebhook(id, "")
}

// UpdateWebhook updates a webhook.
func (c *Core) UpdateWebhook(id int, o models.Webhook) (models.Webhook, error) {
	if o.Events == nil {
		o.Events = pq.StringArray{}
	}

	res, err := c.q.UpdateWebhook.Exec(id, o.Name, o.URL, o.Secret, o.Enabled, pq.StringArray(o.Events), o.MaxRetries, o.TimeoutSeconds)
	if err != nil {
		c.log.Printf("error updating webhook: %v", err)
		return models.Webhook{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "webhook", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.Webhook{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "webhook"))
	}

	return c.GetWebhook(id, "")
}

// DeleteWebhook deletes a webhook.
func (c *Core) DeleteWebhook(id int) error {
	res, err := c.q.DeleteWebhook.Exec(id)
	if err != nil {
		c.log.Printf("error deleting webhook: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "webhook", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "webhook"))
	}

	return nil
}

// GetWebhooksByEvent returns all enabled webhooks matching a given event.
func (c *Core) GetWebhooksByEvent(event string) (models.Webhooks, error) {
	var out models.Webhooks
	if err := c.q.GetWebhooksByEvent.Select(&out, event); err != nil {
		c.log.Printf("error fetching webhooks by event: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "webhooks", "error", pqErrMsg(err)))
	}
	return out, nil
}

// InsertWebhookLog records a webhook delivery attempt.
func (c *Core) InsertWebhookLog(webhookID int, event string, payload []byte, respCode int, respBody, errMsg string, attempt int) error {
	if _, err := c.q.InsertWebhookLog.Exec(webhookID, event, payload, respCode, respBody, errMsg, attempt); err != nil {
		c.log.Printf("error inserting webhook log: %v", err)
		return err
	}
	return nil
}
