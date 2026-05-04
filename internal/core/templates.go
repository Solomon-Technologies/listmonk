package core

import (
	"database/sql"
	"net/http"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	null "gopkg.in/volatiletech/null.v6"
)

// GetTemplates retrieves all templates.
// companyID=0 disables tenant filtering; >0 scopes results.
func (c *Core) GetTemplates(status string, noBody bool, companyID int) ([]models.Template, error) {
	out := []models.Template{}
	if err := c.q.GetTemplates.Select(&out, 0, noBody, status, companyID); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.templates}", "error", pqErrMsg(err)))
	}

	return out, nil
}

// GetTemplate retrieves a given template.
// companyID=0 disables tenant filtering (used for internal flows like
// campaign template merge); >0 scopes the lookup.
func (c *Core) GetTemplate(id int, noBody bool, companyID int) (models.Template, error) {
	var out []models.Template
	if err := c.q.GetTemplates.Select(&out, id, noBody, "", companyID); err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.templates}", "error", pqErrMsg(err)))
	}

	if len(out) == 0 {
		return models.Template{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "{globals.terms.template}"))
	}

	return out[0], nil
}

// CreateTemplate creates a new template.
// companyID stamps the template's tenant; 0 falls back to Solomon=1 in SQL.
func (c *Core) CreateTemplate(name, typ, subject string, body []byte, bodySource null.String, companyID int) (models.Template, error) {
	var newID int
	if err := c.q.CreateTemplate.Get(&newID, name, typ, subject, body, bodySource, companyID); err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "{globals.terms.template}", "error", pqErrMsg(err)))
	}

	return c.GetTemplate(newID, false, 0)
}

// UpdateTemplate updates a given template.
func (c *Core) UpdateTemplate(id int, name, subject string, body []byte, bodySource null.String) (models.Template, error) {
	res, err := c.q.UpdateTemplate.Exec(id, name, subject, body, bodySource)
	if err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.terms.template}", "error", pqErrMsg(err)))
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return models.Template{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "{globals.terms.template}"))
	}

	return c.GetTemplate(id, false, 0)
}

// SetDefaultTemplate sets a template as default.
func (c *Core) SetDefaultTemplate(id int) error {
	if _, err := c.q.SetDefaultTemplate.Exec(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.terms.template}", "error", pqErrMsg(err)))
	}

	return nil
}

// DeleteTemplate deletes a given template.
func (c *Core) DeleteTemplate(id int) error {
	var delID int
	if err := c.q.DeleteTemplate.Get(&delID, id); err != nil && err != sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "{globals.terms.template}", "error", pqErrMsg(err)))
	}
	if delID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, c.i18n.T("templates.cantDeleteDefault"))
	}

	return nil
}
