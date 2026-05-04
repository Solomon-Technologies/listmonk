package main

import (
	"net/http"

	"github.com/knadh/listmonk/internal/auth"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetWebhooks handles retrieval of webhooks with pagination.
func (a *App) GetWebhooks(c echo.Context) error {
	var (
		pg      = a.pg.NewFromURL(c.Request().URL.Query())
		search  = c.FormValue("query")
		orderBy = c.FormValue("order_by")
		order   = c.FormValue("order")
	)

	out, total, err := a.core.QueryWebhooks(search, orderBy, order, pg.Offset, pg.Limit, a.tenantFilter(c))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Results models.Webhooks `json:"results"`
		Total   int             `json:"total"`
		Page    int             `json:"page"`
		PerPage int             `json:"per_page"`
	}{out, total, pg.Page, pg.PerPage}})
}

// GetWebhook handles retrieval of a single webhook.
func (a *App) GetWebhook(c echo.Context) error {
	id := getID(c)

	out, err := a.core.GetWebhook(id, "", a.tenantFilter(c))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateWebhook handles creation of a new webhook.
func (a *App) CreateWebhook(c echo.Context) error {
	var o models.Webhook
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	if o.Name == "" || o.URL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and url are required")
	}

	// Validate events.
	for _, e := range o.Events {
		if !models.WebhookEvents[e] {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid event: "+e)
		}
	}

	if o.MaxRetries == 0 {
		o.MaxRetries = 3
	}
	if o.TimeoutSeconds == 0 {
		o.TimeoutSeconds = 10
	}

	user := auth.GetUser(c)
	out, err := a.core.CreateWebhook(o, user.CompanyID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateWebhook handles update of a webhook.
func (a *App) UpdateWebhook(c echo.Context) error {
	id := getID(c)

	var o models.Webhook
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	// Validate events.
	for _, e := range o.Events {
		if !models.WebhookEvents[e] {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid event: "+e)
		}
	}

	out, err := a.core.UpdateWebhook(id, o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteWebhook handles deletion of a webhook.
func (a *App) DeleteWebhook(c echo.Context) error {
	id := getID(c)

	if err := a.core.DeleteWebhook(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// GetWebhookLog handles retrieval of webhook delivery logs.
func (a *App) GetWebhookLog(c echo.Context) error {
	id := getID(c)
	pg := a.pg.NewFromURL(c.Request().URL.Query())

	var out models.WebhookLogs
	if err := a.db.Select(&out, a.queries.QueryWebhookLog, id, pg.Offset, pg.Limit); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error fetching webhook log")
	}

	total := 0
	if len(out) > 0 {
		total = out[0].Total
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Results models.WebhookLogs `json:"results"`
		Total   int                `json:"total"`
		Page    int                `json:"page"`
		PerPage int                `json:"per_page"`
	}{out, total, pg.Page, pg.PerPage}})
}

// TestWebhook sends a test payload to a webhook endpoint.
func (a *App) TestWebhook(c echo.Context) error {
	id := getID(c)

	wh, err := a.core.GetWebhook(id, "", a.tenantFilter(c))
	if err != nil {
		return err
	}

	// Send a test event via the webhook manager if available.
	if a.webhookMgr != nil {
		a.webhookMgr.Dispatch("webhook.test", map[string]any{
			"webhook_id": wh.ID,
			"test":       true,
		})
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}
