package core

import (
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetABTest retrieves an A/B test by ID or UUID.
// companyID=0 disables tenant filtering.
func (c *Core) GetABTest(id int, uuStr string, companyID int) (models.ABTest, error) {
	var uu any
	if uuStr != "" {
		uu = uuStr
	}

	var out models.ABTest
	if err := c.q.GetABTest.Get(&out, id, uu, companyID); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "A/B test", "error", pqErrMsg(err)))
	}

	if out.ID == 0 {
		return out, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "A/B test"))
	}

	// Load variants.
	variants, err := c.GetABVariants(out.ID)
	if err != nil {
		return out, err
	}

	// Compute rates.
	for i := range variants {
		if variants[i].Sent > 0 {
			variants[i].OpenRate = float64(variants[i].Opened) / float64(variants[i].Sent) * 100
			variants[i].ClickRate = float64(variants[i].Clicked) / float64(variants[i].Sent) * 100
		}
	}
	out.Variants = variants

	return out, nil
}

// GetABTestByCampaign retrieves the A/B test for a campaign.
// companyID=0 disables tenant filtering.
func (c *Core) GetABTestByCampaign(campaignID, companyID int) (models.ABTest, error) {
	var out models.ABTest
	if err := c.q.GetABTestByCampaign.Get(&out, campaignID, companyID); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "A/B test", "error", pqErrMsg(err)))
	}
	return out, nil
}

// CreateABTest creates a new A/B test. companyID stamps tenant.
func (c *Core) CreateABTest(o models.ABTest, companyID int) (models.ABTest, error) {
	uu, err := uuid.NewV4()
	if err != nil {
		return models.ABTest{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "A/B test", "error", err.Error()))
	}

	if o.Status == "" {
		o.Status = models.ABTestStatusDraft
	}
	if o.TestType == "" {
		o.TestType = models.ABTestTypeSubject
	}
	if o.TestPercentage == 0 {
		o.TestPercentage = 20
	}
	if o.WinnerMetric == "" {
		o.WinnerMetric = models.ABMetricOpenRate
	}
	if o.WinnerWaitHours == 0 {
		o.WinnerWaitHours = 4
	}

	var id int
	if err := c.q.CreateABTest.Get(&id, uu, o.CampaignID, o.TestType, o.Status, o.TestPercentage, o.WinnerMetric, o.WinnerWaitHours, companyID); err != nil {
		c.log.Printf("error creating A/B test: %v", err)
		return models.ABTest{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "A/B test", "error", pqErrMsg(err)))
	}

	return c.GetABTest(id, "", 0)
}

// UpdateABTest updates an A/B test.
func (c *Core) UpdateABTest(id int, o models.ABTest) (models.ABTest, error) {
	res, err := c.q.UpdateABTest.Exec(id, o.TestType, o.TestPercentage, o.WinnerMetric, o.WinnerWaitHours)
	if err != nil {
		c.log.Printf("error updating A/B test: %v", err)
		return models.ABTest{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "A/B test", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.ABTest{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "A/B test"))
	}

	return c.GetABTest(id, "", 0)
}

// UpdateABTestStatus updates the status of an A/B test.
func (c *Core) UpdateABTestStatus(id int, status string, winnerID int) error {
	_, err := c.q.UpdateABTestStatus.Exec(id, status, winnerID)
	if err != nil {
		c.log.Printf("error updating A/B test status: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "A/B test", "error", pqErrMsg(err)))
	}
	return nil
}

// DeleteABTest deletes an A/B test.
func (c *Core) DeleteABTest(id int) error {
	res, err := c.q.DeleteABTest.Exec(id)
	if err != nil {
		c.log.Printf("error deleting A/B test: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "A/B test", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "A/B test"))
	}
	return nil
}

// GetABVariants returns all variants for an A/B test.
func (c *Core) GetABVariants(abTestID int) ([]models.ABTestVariant, error) {
	var out []models.ABTestVariant
	if err := c.q.GetABVariants.Select(&out, abTestID); err != nil {
		c.log.Printf("error fetching A/B variants: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "A/B variants", "error", pqErrMsg(err)))
	}
	return out, nil
}

// CreateABVariant creates a new A/B test variant.
func (c *Core) CreateABVariant(o models.ABTestVariant) (models.ABTestVariant, error) {
	var id int
	if err := c.q.CreateABVariant.Get(&id, o.ABTestID, o.Label, o.Subject, o.Body, o.FromEmail); err != nil {
		c.log.Printf("error creating A/B variant: %v", err)
		return models.ABTestVariant{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "A/B variant", "error", pqErrMsg(err)))
	}

	var out models.ABTestVariant
	if err := c.q.GetABVariant.Get(&out, id); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "A/B variant", "error", pqErrMsg(err)))
	}
	return out, nil
}

// UpdateABVariant updates a variant.
func (c *Core) UpdateABVariant(id int, o models.ABTestVariant) (models.ABTestVariant, error) {
	res, err := c.q.UpdateABVariant.Exec(id, o.Label, o.Subject, o.Body, o.FromEmail)
	if err != nil {
		c.log.Printf("error updating A/B variant: %v", err)
		return models.ABTestVariant{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "A/B variant", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return models.ABTestVariant{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "A/B variant"))
	}

	var out models.ABTestVariant
	if err := c.q.GetABVariant.Get(&out, id); err != nil {
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "A/B variant", "error", pqErrMsg(err)))
	}
	return out, nil
}

// DeleteABVariant deletes a variant.
func (c *Core) DeleteABVariant(id int) error {
	res, err := c.q.DeleteABVariant.Exec(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "A/B variant", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "A/B variant"))
	}
	return nil
}

// GetRunningABTests returns A/B tests that are running and past their wait period.
func (c *Core) GetRunningABTests() ([]models.ABTest, error) {
	var out []models.ABTest
	if err := c.q.GetRunningABTests.Select(&out); err != nil {
		c.log.Printf("error fetching running A/B tests: %v", err)
		return nil, err
	}
	return out, nil
}
