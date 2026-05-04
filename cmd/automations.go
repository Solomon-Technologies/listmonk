package main

import (
	"net/http"
	"strconv"

	"github.com/knadh/listmonk/internal/auth"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetAutomations handles retrieval of automations with pagination.
func (a *App) GetAutomations(c echo.Context) error {
	var (
		pg      = a.pg.NewFromURL(c.Request().URL.Query())
		search  = c.FormValue("query")
		orderBy = c.FormValue("order_by")
		order   = c.FormValue("order")
	)

	out, total, err := a.core.QueryAutomations(search, orderBy, order, pg.Offset, pg.Limit, a.tenantFilter(c))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Results models.Automations `json:"results"`
		Total   int                `json:"total"`
		Page    int                `json:"page"`
		PerPage int                `json:"per_page"`
	}{out, total, pg.Page, pg.PerPage}})
}

// GetAutomation handles retrieval of a single automation with nodes and edges.
func (a *App) GetAutomation(c echo.Context) error {
	id := getID(c)

	out, err := a.core.GetAutomation(id, "", a.tenantFilter(c))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateAutomation handles creation of a new automation.
func (a *App) CreateAutomation(c echo.Context) error {
	var o models.Automation
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	if o.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	user := auth.GetUser(c)
	out, err := a.core.CreateAutomation(o, user.CompanyID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateAutomation handles update of an automation.
func (a *App) UpdateAutomation(c echo.Context) error {
	id := getID(c)

	var o models.Automation
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	out, err := a.core.UpdateAutomation(id, o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteAutomation handles deletion of an automation.
func (a *App) DeleteAutomation(c echo.Context) error {
	id := getID(c)

	if err := a.core.DeleteAutomation(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// CreateAutomationNode handles creation of a node.
func (a *App) CreateAutomationNode(c echo.Context) error {
	automationID := getID(c)

	var o models.AutomationNode
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	o.AutomationID = automationID

	if !models.AutomationNodeTypes[o.NodeType] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_type: "+o.NodeType)
	}

	out, err := a.core.CreateAutomationNode(o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateAutomationNode handles update of a node.
func (a *App) UpdateAutomationNode(c echo.Context) error {
	nodeID, _ := strconv.Atoi(c.Param("nodeID"))
	if nodeID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node ID")
	}

	var o models.AutomationNode
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	out, err := a.core.UpdateAutomationNode(nodeID, o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteAutomationNode handles deletion of a node.
func (a *App) DeleteAutomationNode(c echo.Context) error {
	nodeID, _ := strconv.Atoi(c.Param("nodeID"))
	if nodeID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node ID")
	}

	if err := a.core.DeleteAutomationNode(nodeID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// CreateAutomationEdge handles creation of an edge.
func (a *App) CreateAutomationEdge(c echo.Context) error {
	automationID := getID(c)

	var o models.AutomationEdge
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	o.AutomationID = automationID

	out, err := a.core.CreateAutomationEdge(o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteAutomationEdge handles deletion of an edge.
func (a *App) DeleteAutomationEdge(c echo.Context) error {
	edgeID, _ := strconv.Atoi(c.Param("edgeID"))
	if edgeID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid edge ID")
	}

	if err := a.core.DeleteAutomationEdge(edgeID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}
