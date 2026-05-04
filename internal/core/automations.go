package core

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

var automationQuerySortFields = []string{"name", "status", "created_at", "updated_at"}

// QueryAutomations retrieves paginated automations.
// companyID=0 disables tenant filtering.
func (c *Core) QueryAutomations(searchStr, orderBy, order string, offset, limit, companyID int) (models.Automations, int, error) {
	queryStr, stmt := makeSearchQuery(searchStr, orderBy, order, c.q.QueryAutomations, automationQuerySortFields)

	var out models.Automations
	if err := c.db.Select(&out, stmt, 0, nil, queryStr, offset, limit, companyID); err != nil {
		c.log.Printf("error fetching automations: %v", err)
		return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "automations", "error", pqErrMsg(err)))
	}

	total := 0
	if len(out) > 0 {
		total = out[0].Total
	}

	return out, total, nil
}

// GetAutomation retrieves an automation by ID or UUID with nodes and edges.
// companyID=0 disables tenant filtering.
func (c *Core) GetAutomation(id int, uuStr string, companyID int) (models.Automation, error) {
	var uu any
	if uuStr != "" {
		uu = uuStr
	}

	var out models.Automation
	if err := c.q.GetAutomation.Get(&out, id, uu, companyID); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "automation", "error", pqErrMsg(err)))
	}

	if out.ID == 0 {
		return out, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "automation"))
	}

	// Load nodes and edges.
	nodes, _ := c.GetAutomationNodes(out.ID)
	edges, _ := c.GetAutomationEdges(out.ID)
	out.Nodes = nodes
	out.Edges = edges

	return out, nil
}

// CreateAutomation creates a new automation. companyID stamps tenant.
func (c *Core) CreateAutomation(o models.Automation, companyID int) (models.Automation, error) {
	uu, err := uuid.NewV4()
	if err != nil {
		return models.Automation{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "automation", "error", err.Error()))
	}

	if o.Status == "" {
		o.Status = models.AutomationStatusDraft
	}
	if o.Canvas == nil {
		o.Canvas = []byte(`{"nodes":[],"edges":[]}`)
	}

	var id int
	if err := c.q.CreateAutomation.Get(&id, uu, o.Name, o.Description, o.Status, o.Canvas, companyID); err != nil {
		c.log.Printf("error creating automation: %v", err)
		return models.Automation{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "automation", "error", pqErrMsg(err)))
	}

	return c.GetAutomation(id, "", 0)
}

// UpdateAutomation updates an automation.
func (c *Core) UpdateAutomation(id int, o models.Automation) (models.Automation, error) {
	if o.Canvas == nil {
		o.Canvas = []byte(`{"nodes":[],"edges":[]}`)
	}

	res, err := c.q.UpdateAutomation.Exec(id, o.Name, o.Description, o.Status, o.Canvas)
	if err != nil {
		c.log.Printf("error updating automation: %v", err)
		return models.Automation{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "automation", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.Automation{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "automation"))
	}

	return c.GetAutomation(id, "", 0)
}

// DeleteAutomation deletes an automation.
func (c *Core) DeleteAutomation(id int) error {
	res, err := c.q.DeleteAutomation.Exec(id)
	if err != nil {
		c.log.Printf("error deleting automation: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "automation", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "automation"))
	}
	return nil
}

// GetAutomationNodes returns all nodes for an automation.
func (c *Core) GetAutomationNodes(automationID int) ([]models.AutomationNode, error) {
	var out []models.AutomationNode
	if err := c.q.GetAutomationNodes.Select(&out, automationID); err != nil {
		c.log.Printf("error fetching automation nodes: %v", err)
		return nil, err
	}
	return out, nil
}

// GetAutomationEdges returns all edges for an automation.
func (c *Core) GetAutomationEdges(automationID int) ([]models.AutomationEdge, error) {
	var out []models.AutomationEdge
	if err := c.q.GetAutomationEdges.Select(&out, automationID); err != nil {
		c.log.Printf("error fetching automation edges: %v", err)
		return nil, err
	}
	return out, nil
}

// CreateAutomationNode creates a new node.
func (c *Core) CreateAutomationNode(o models.AutomationNode) (models.AutomationNode, error) {
	uu, err := uuid.NewV4()
	if err != nil {
		return models.AutomationNode{}, err
	}

	if o.Config == nil {
		o.Config = []byte("{}")
	}

	var id int
	if err := c.q.CreateAutomationNode.Get(&id, uu, o.AutomationID, o.NodeType, o.Config, o.PositionX, o.PositionY); err != nil {
		c.log.Printf("error creating automation node: %v", err)
		return models.AutomationNode{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "automation node", "error", pqErrMsg(err)))
	}

	var out models.AutomationNode
	c.q.GetAutomationNode.Get(&out, id)
	return out, nil
}

// UpdateAutomationNode updates a node.
func (c *Core) UpdateAutomationNode(id int, o models.AutomationNode) (models.AutomationNode, error) {
	if o.Config == nil {
		o.Config = []byte("{}")
	}

	res, err := c.q.UpdateAutomationNode.Exec(id, o.NodeType, o.Config, o.PositionX, o.PositionY)
	if err != nil {
		return models.AutomationNode{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "automation node", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.AutomationNode{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "automation node"))
	}

	var out models.AutomationNode
	c.q.GetAutomationNode.Get(&out, id)
	return out, nil
}

// DeleteAutomationNode deletes a node.
func (c *Core) DeleteAutomationNode(id int) error {
	_, err := c.q.DeleteAutomationNode.Exec(id)
	return err
}

// CreateAutomationEdge creates a connection between two nodes.
func (c *Core) CreateAutomationEdge(o models.AutomationEdge) (models.AutomationEdge, error) {
	var id int
	if err := c.q.CreateAutomationEdge.Get(&id, o.AutomationID, o.FromNodeID, o.ToNodeID, o.Label); err != nil {
		return models.AutomationEdge{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "automation edge", "error", pqErrMsg(err)))
	}
	o.ID = id
	return o, nil
}

// DeleteAutomationEdge deletes an edge.
func (c *Core) DeleteAutomationEdge(id int) error {
	_, err := c.q.DeleteAutomationEdge.Exec(id)
	return err
}
