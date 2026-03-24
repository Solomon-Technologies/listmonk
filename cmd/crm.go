package main

import (
	"net/http"
	"strconv"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetDeals handles retrieval of deals with pagination.
func (a *App) GetDeals(c echo.Context) error {
	pg := a.pg.NewFromURL(c.Request().URL.Query())
	subscriberID, _ := strconv.Atoi(c.FormValue("subscriber_id"))
	status := c.FormValue("status")

	out, total, err := a.core.QueryDeals(subscriberID, status, pg.Offset, pg.Limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Results models.Deals `json:"results"`
		Total   int          `json:"total"`
		Page    int          `json:"page"`
		PerPage int          `json:"per_page"`
	}{out, total, pg.Page, pg.PerPage}})
}

// GetDeal handles retrieval of a single deal.
func (a *App) GetDeal(c echo.Context) error {
	id := getID(c)

	out, err := a.core.GetDeal(id, "")
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateDeal handles creation of a new deal.
func (a *App) CreateDeal(c echo.Context) error {
	var o models.Deal
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	if o.Name == "" || o.SubscriberID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "name and subscriber_id are required")
	}

	out, err := a.core.CreateDeal(o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateDeal handles update of a deal.
func (a *App) UpdateDeal(c echo.Context) error {
	id := getID(c)

	var o models.Deal
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	out, err := a.core.UpdateDeal(id, o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteDeal handles deletion of a deal.
func (a *App) DeleteDeal(c echo.Context) error {
	id := getID(c)

	if err := a.core.DeleteDeal(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// GetDealPipeline handles retrieval of the deal pipeline summary.
func (a *App) GetDealPipeline(c echo.Context) error {
	out, err := a.core.GetDealPipeline()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// GetSubscriberActivities handles retrieval of activities for a subscriber.
func (a *App) GetSubscriberActivities(c echo.Context) error {
	id := getID(c)
	pg := a.pg.NewFromURL(c.Request().URL.Query())

	out, total, err := a.core.GetSubscriberActivities(id, pg.Offset, pg.Limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Results models.ContactActivities `json:"results"`
		Total   int                      `json:"total"`
		Page    int                      `json:"page"`
		PerPage int                      `json:"per_page"`
	}{out, total, pg.Page, pg.PerPage}})
}

// CreateActivity handles creation of a contact activity.
func (a *App) CreateActivity(c echo.Context) error {
	subscriberID := getID(c)

	var o models.ContactActivity
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	o.SubscriberID = subscriberID

	if o.ActivityType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "activity_type is required")
	}

	if !models.ActivityTypes[o.ActivityType] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid activity_type: "+o.ActivityType)
	}

	out, err := a.core.CreateContactActivity(o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteActivity handles deletion of a contact activity.
func (a *App) DeleteActivity(c echo.Context) error {
	activityID, _ := strconv.ParseInt(c.Param("activityID"), 10, 64)
	if activityID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid activity ID")
	}

	if err := a.core.DeleteActivity(activityID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}
