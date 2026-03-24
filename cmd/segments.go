package main

import (
	"net/http"

	"github.com/knadh/listmonk/internal/core"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetSegments handles retrieval of segments with pagination and filters.
func (a *App) GetSegments(c echo.Context) error {
	var (
		pg     = a.pg.NewFromURL(c.Request().URL.Query())
		search = c.FormValue("query")
		tags   = c.QueryParams()["tag"]

		orderBy = c.FormValue("order_by")
		order   = c.FormValue("order")
	)

	out, total, err := a.core.QuerySegments(search, tags, orderBy, order, pg.Offset, pg.Limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Results models.Segments `json:"results"`
		Total   int             `json:"total"`
		Page    int             `json:"page"`
		PerPage int             `json:"per_page"`
	}{out, total, pg.Page, pg.PerPage}})
}

// GetSegment handles retrieval of a single segment.
func (a *App) GetSegment(c echo.Context) error {
	id := getID(c)

	out, err := a.core.GetSegment(id, "")
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateSegment handles creation of a new segment.
func (a *App) CreateSegment(c echo.Context) error {
	var o models.Segment
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	if o.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	if o.MatchType == "" {
		o.MatchType = models.SegmentMatchAll
	}

	out, err := a.core.CreateSegment(o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateSegment handles update of a segment.
func (a *App) UpdateSegment(c echo.Context) error {
	id := getID(c)

	var o models.Segment
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	out, err := a.core.UpdateSegment(id, o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteSegment handles deletion of a segment.
func (a *App) DeleteSegment(c echo.Context) error {
	id := getID(c)

	if err := a.core.DeleteSegment(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// GetSegmentCount returns the dynamic subscriber count for a segment.
func (a *App) GetSegmentCount(c echo.Context) error {
	id := getID(c)

	seg, err := a.core.GetSegment(id, "")
	if err != nil {
		return err
	}

	count, err := a.core.GetSegmentSubscriberCount(seg)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Count int `json:"count"`
	}{count}})
}

// GetSegmentSubscribers returns paginated subscribers matching a segment.
func (a *App) GetSegmentSubscribers(c echo.Context) error {
	id := getID(c)

	seg, err := a.core.GetSegment(id, "")
	if err != nil {
		return err
	}

	// Build WHERE clause from segment conditions.
	queryExp, err := core.BuildSegmentWhere(seg.Conditions, seg.MatchType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid segment conditions: "+err.Error())
	}

	pg := a.pg.NewFromURL(c.Request().URL.Query())

	// Query subscribers with the segment's WHERE clause using the existing infrastructure.
	out, total, err := a.core.QuerySubscribers("", queryExp, nil, "", "id", "asc", pg.Offset, pg.Limit)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Results any `json:"results"`
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	}{out, total, pg.Page, pg.PerPage}})
}
