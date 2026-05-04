package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// GetCompanies returns the full tenant list. Read access required so that
// the user-create and role-create dropdowns can populate company pickers.
func (a *App) GetCompanies(c echo.Context) error {
	out, err := a.core.GetCompanies()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: out})
}

// GetCompany returns a single tenant by id.
func (a *App) GetCompany(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	out, err := a.core.GetCompany(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: out})
}

// GetCompanyStats returns each tenant + counts of its dependent rows.
// Powers the Companies admin page so destructive deletes can show a
// "blocked because X rows reference this" preview.
func (a *App) GetCompanyStats(c echo.Context) error {
	out, err := a.core.GetCompanyStats()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateCompany handles tenant creation.
func (a *App) CreateCompany(c echo.Context) error {
	var o struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.Bind(&o); err != nil {
		return err
	}
	o.Name = strings.TrimSpace(o.Name)
	o.Slug = strings.TrimSpace(o.Slug)
	if o.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if o.Slug == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "slug is required")
	}

	out, err := a.core.CreateCompany(o.Name, o.Slug)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateCompany handles tenant edit.
func (a *App) UpdateCompany(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var o struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.Bind(&o); err != nil {
		return err
	}

	out, err := a.core.UpdateCompany(id, strings.TrimSpace(o.Name), strings.TrimSpace(o.Slug))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteCompany removes a tenant. Fails (with friendly message) when any
// tenant data still references the company_id.
func (a *App) DeleteCompany(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := a.core.DeleteCompany(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: true})
}
