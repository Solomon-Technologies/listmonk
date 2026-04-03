package main

import (
	"net/http"
	"strconv"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetDripCampaigns handles retrieval of drip campaigns with pagination.
func (a *App) GetDripCampaigns(c echo.Context) error {
	var (
		pg      = a.pg.NewFromURL(c.Request().URL.Query())
		search  = c.FormValue("query")
		orderBy = c.FormValue("order_by")
		order   = c.FormValue("order")
	)

	out, total, err := a.core.QueryDripCampaigns(search, orderBy, order, pg.Offset, pg.Limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Results models.DripCampaigns `json:"results"`
		Total   int                  `json:"total"`
		Page    int                  `json:"page"`
		PerPage int                  `json:"per_page"`
	}{out, total, pg.Page, pg.PerPage}})
}

// GetDripCampaign handles retrieval of a single drip campaign with steps.
func (a *App) GetDripCampaign(c echo.Context) error {
	id := getID(c)

	out, err := a.core.GetDripCampaign(id, "")
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateDripCampaign handles creation of a new drip campaign.
func (a *App) CreateDripCampaign(c echo.Context) error {
	var o models.DripCampaign
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	if o.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	out, err := a.core.CreateDripCampaign(o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateDripCampaign handles update of a drip campaign.
func (a *App) UpdateDripCampaign(c echo.Context) error {
	id := getID(c)

	var o models.DripCampaign
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	out, err := a.core.UpdateDripCampaign(id, o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateDripCampaignStatus handles status change (activate, pause, archive).
func (a *App) UpdateDripCampaignStatus(c echo.Context) error {
	id := getID(c)

	var o struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	if !models.DripStatuses[o.Status] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status: "+o.Status)
	}

	if err := a.core.UpdateDripCampaignStatus(id, o.Status); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// DeleteDripCampaign handles deletion of a drip campaign.
func (a *App) DeleteDripCampaign(c echo.Context) error {
	id := getID(c)

	if err := a.core.DeleteDripCampaign(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// GetDripSteps handles retrieval of steps for a drip campaign.
func (a *App) GetDripSteps(c echo.Context) error {
	id := getID(c)

	out, err := a.core.GetDripSteps(id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateDripStep handles creation of a new drip step.
func (a *App) CreateDripStep(c echo.Context) error {
	dripID := getID(c)

	var o models.DripStep
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	o.DripCampaignID = dripID

	if o.Subject == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "subject is required")
	}

	out, err := a.core.CreateDripStep(o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateDripStep handles update of a drip step.
func (a *App) UpdateDripStep(c echo.Context) error {
	stepID, _ := strconv.Atoi(c.Param("stepID"))
	if stepID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid step ID")
	}

	var o models.DripStep
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	out, err := a.core.UpdateDripStep(stepID, o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteDripStep handles deletion of a drip step.
func (a *App) DeleteDripStep(c echo.Context) error {
	stepID, _ := strconv.Atoi(c.Param("stepID"))
	if stepID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid step ID")
	}

	if err := a.core.DeleteDripStep(stepID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// GetDripEnrollments handles retrieval of enrollments for a drip campaign.
func (a *App) GetDripEnrollments(c echo.Context) error {
	id := getID(c)
	pg := a.pg.NewFromURL(c.Request().URL.Query())

	out, total, err := a.core.GetDripEnrollments(id, pg.Offset, pg.Limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Results models.DripEnrollments `json:"results"`
		Total   int                    `json:"total"`
		Page    int                    `json:"page"`
		PerPage int                    `json:"per_page"`
	}{out, total, pg.Page, pg.PerPage}})
}

// EnrollSubscriberInDrip handles manual enrollment of a subscriber.
func (a *App) EnrollSubscriberInDrip(c echo.Context) error {
	id := getID(c)

	var o struct {
		SubscriberID int `json:"subscriber_id"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	if o.SubscriberID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "subscriber_id is required")
	}

	if err := a.core.EnrollSubscriberInDrip(id, o.SubscriberID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// BulkEnrollInDrip handles bulk enrollment of multiple subscribers.
func (a *App) BulkEnrollInDrip(c echo.Context) error {
	id := getID(c)

	var o struct {
		SubscriberIDs []int `json:"subscriber_ids"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	if len(o.SubscriberIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "subscriber_ids is required")
	}

	n, err := a.core.BulkEnrollInDrip(id, o.SubscriberIDs)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Enrolled int `json:"enrolled"`
	}{n}})
}
