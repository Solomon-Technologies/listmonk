package core

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

// reSlug matches a valid URL-safe slug: lowercase letters, digits, hyphens.
var reSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// GetCompanies returns all tenants.
func (c *Core) GetCompanies() ([]models.Company, error) {
	out := []models.Company{}
	if err := c.q.GetCompanies.Select(&out); err != nil {
		c.log.Printf("error fetching companies: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "companies", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetCompany returns a single tenant by id.
func (c *Core) GetCompany(id int) (models.Company, error) {
	var out models.Company
	if err := c.q.GetCompany.Get(&out, id); err != nil {
		c.log.Printf("error fetching company %d: %v", id, err)
		return out, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "company"))
	}
	return out, nil
}

// GetCompanyStats returns each tenant with row counts for its dependent
// tables. Used by the admin Companies page so the operator knows what
// referencing data exists before attempting a delete.
func (c *Core) GetCompanyStats() ([]models.CompanyStats, error) {
	out := []models.CompanyStats{}
	if err := c.q.GetCompanyStats.Select(&out); err != nil {
		c.log.Printf("error fetching company stats: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "company stats", "error", pqErrMsg(err)))
	}
	return out, nil
}

// CreateCompany creates a new tenant.
func (c *Core) CreateCompany(name, slug string) (models.Company, error) {
	name = strings.TrimSpace(name)
	slug = strings.ToLower(strings.TrimSpace(slug))
	if name == "" {
		return models.Company{}, echo.NewHTTPError(http.StatusBadRequest, "company name is required")
	}
	if slug == "" {
		return models.Company{}, echo.NewHTTPError(http.StatusBadRequest, "company slug is required")
	}
	if !reSlug.MatchString(slug) {
		return models.Company{}, echo.NewHTTPError(http.StatusBadRequest,
			"slug must be lowercase letters, digits, and hyphens (e.g. 'acme-co')")
	}

	var out models.Company
	if err := c.q.CreateCompany.Get(&out, name, slug); err != nil {
		// Unique-violation on name or slug.
		if pq, ok := err.(*pq.Error); ok && pq.Code == "23505" {
			return models.Company{}, echo.NewHTTPError(http.StatusBadRequest,
				"a company with that name or slug already exists")
		}
		c.log.Printf("error creating company: %v", err)
		return models.Company{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "company", "error", pqErrMsg(err)))
	}
	return out, nil
}

// UpdateCompany updates a tenant's name and/or slug.
func (c *Core) UpdateCompany(id int, name, slug string) (models.Company, error) {
	name = strings.TrimSpace(name)
	slug = strings.ToLower(strings.TrimSpace(slug))
	if slug != "" && !reSlug.MatchString(slug) {
		return models.Company{}, echo.NewHTTPError(http.StatusBadRequest,
			"slug must be lowercase letters, digits, and hyphens")
	}

	var out models.Company
	if err := c.q.UpdateCompany.Get(&out, id, name, slug); err != nil {
		if pq, ok := err.(*pq.Error); ok && pq.Code == "23505" {
			return models.Company{}, echo.NewHTTPError(http.StatusBadRequest,
				"a company with that name or slug already exists")
		}
		c.log.Printf("error updating company %d: %v", id, err)
		return models.Company{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "company", "error", pqErrMsg(err)))
	}
	return out, nil
}

// DeleteCompany removes a tenant. The schema has ON DELETE RESTRICT on every
// company_id FK, so this fails if any tenant data still references it.
func (c *Core) DeleteCompany(id int) error {
	res, err := c.q.DeleteCompany.Exec(id)
	if err != nil {
		// 23503 = foreign_key_violation. Surface a friendly message.
		if pq, ok := err.(*pq.Error); ok && pq.Code == "23503" {
			return echo.NewHTTPError(http.StatusBadRequest,
				"cannot delete this company while it has lists, subscribers, campaigns, users, or other tenant data referencing it. Reassign or delete those first.")
		}
		c.log.Printf("error deleting company %d: %v", id, err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "company", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "company"))
	}
	return nil
}
