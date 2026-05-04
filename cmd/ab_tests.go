package main

import (
	"net/http"
	"strconv"

	"github.com/knadh/listmonk/internal/auth"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetABTest handles retrieval of an A/B test.
func (a *App) GetABTest(c echo.Context) error {
	id := getID(c)

	out, err := a.core.GetABTest(id, "", a.tenantFilter(c))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// GetABTestByCampaign handles retrieval of an A/B test for a campaign.
func (a *App) GetABTestByCampaign(c echo.Context) error {
	id := getID(c)

	out, err := a.core.GetABTestByCampaign(id, a.tenantFilter(c))
	if err != nil {
		return err
	}

	// Load variants.
	variants, err := a.core.GetABVariants(out.ID)
	if err != nil {
		return err
	}
	for i := range variants {
		if variants[i].Sent > 0 {
			variants[i].OpenRate = float64(variants[i].Opened) / float64(variants[i].Sent) * 100
			variants[i].ClickRate = float64(variants[i].Clicked) / float64(variants[i].Sent) * 100
		}
	}
	out.Variants = variants

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateABTest handles creation of a new A/B test.
func (a *App) CreateABTest(c echo.Context) error {
	var o models.ABTest
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	if o.CampaignID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "campaign_id is required")
	}

	if o.TestType != "" && !models.ABTestTypes[o.TestType] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid test_type: "+o.TestType)
	}

	user := auth.GetUser(c)
	out, err := a.core.CreateABTest(o, user.CompanyID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateABTest handles update of an A/B test.
func (a *App) UpdateABTest(c echo.Context) error {
	id := getID(c)

	var o models.ABTest
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	out, err := a.core.UpdateABTest(id, o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteABTest handles deletion of an A/B test.
func (a *App) DeleteABTest(c echo.Context) error {
	id := getID(c)

	if err := a.core.DeleteABTest(id); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}

// CreateABVariant handles creation of a new variant.
func (a *App) CreateABVariant(c echo.Context) error {
	abTestID := getID(c)

	var o models.ABTestVariant
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	o.ABTestID = abTestID

	out, err := a.core.CreateABVariant(o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateABVariant handles update of a variant.
func (a *App) UpdateABVariant(c echo.Context) error {
	variantID, _ := strconv.Atoi(c.Param("variantID"))
	if variantID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid variant ID")
	}

	var o models.ABTestVariant
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request: "+err.Error())
	}

	out, err := a.core.UpdateABVariant(variantID, o)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: out})
}

// DeleteABVariant handles deletion of a variant.
func (a *App) DeleteABVariant(c echo.Context) error {
	variantID, _ := strconv.Atoi(c.Param("variantID"))
	if variantID < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid variant ID")
	}

	if err := a.core.DeleteABVariant(variantID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: true})
}
