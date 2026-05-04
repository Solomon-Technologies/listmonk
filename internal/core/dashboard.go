package core

import (
	"net/http"

	"github.com/jmoiron/sqlx/types"
	"github.com/labstack/echo/v4"
)

// GetDashboardCharts returns chart data points to render on the dashboard.
// companyID=0 disables tenant filtering (platform admin sees global aggregate).
// As of v7.17.0 this is on-the-fly (no longer reads mat_dashboard_charts) so
// per-tenant scoping is honored. The mat view itself is still refreshed by
// other code paths but bypassed here.
func (c *Core) GetDashboardCharts(companyID int) (types.JSONText, error) {
	var out types.JSONText
	if err := c.q.GetDashboardCharts.Get(&out, companyID); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "dashboard charts", "error", pqErrMsg(err)))
	}

	return out, nil
}

// GetDashboardCounts returns stats counts to show on the dashboard.
// companyID=0 disables tenant filtering.
func (c *Core) GetDashboardCounts(companyID int) (types.JSONText, error) {
	var out types.JSONText
	if err := c.q.GetDashboardCounts.Get(&out, companyID); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "dashboard stats", "error", pqErrMsg(err)))
	}

	return out, nil
}

// GetDashboardFeatureCounts returns counts for Solomon platform features.
// companyID=0 disables tenant filtering.
func (c *Core) GetDashboardFeatureCounts(companyID int) (types.JSONText, error) {
	var out types.JSONText
	if err := c.q.GetDashboardFeatureCounts.Get(&out, companyID); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "feature counts", "error", pqErrMsg(err)))
	}

	return out, nil
}
