package main

import (
	"net/http"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetScoringRules handles retrieval of all scoring rules.
func (a *App) GetScoringRules(c echo.Context) error {
	out, err := a.core.GetScoringRules()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// GetScoringRule handles retrieval of a single scoring rule.
func (a *App) GetScoringRule(c echo.Context) error {
	id := getID(c)

	out, err := a.core.GetScoringRule(id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateScoringRule handles creation of a scoring rule.
func (a *App) CreateScoringRule(c echo.Context) error {
	var o models.ScoringRule
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	if o.Name == "" || o.EventType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and event_type are required")
	}

	if !models.ScoringEvents[o.EventType] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid event_type: "+o.EventType)
	}

	out, err := a.core.CreateScoringRule(o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateScoringRule handles update of a scoring rule.
func (a *App) UpdateScoringRule(c echo.Context) error {
	id := getID(c)

	var o models.ScoringRule
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	out, err := a.core.UpdateScoringRule(id, o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteScoringRule handles deletion of a scoring rule.
func (a *App) DeleteScoringRule(c echo.Context) error {
	id := getID(c)

	if err := a.core.DeleteScoringRule(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// GetSubscriberScoreLog handles retrieval of score history for a subscriber.
func (a *App) GetSubscriberScoreLog(c echo.Context) error {
	id := getID(c)
	pg := a.pg.NewFromURL(c.Request().URL.Query())

	out, err := a.core.GetSubscriberScoreLog(id, pg.Offset, pg.Limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}
