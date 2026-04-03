package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	null "gopkg.in/volatiletech/null.v6"
)

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

// GetWarmingAddresses returns all warming addresses.
func (a *App) GetWarmingAddresses(c echo.Context) error {
	out, err := a.core.GetWarmingAddresses()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateWarmingAddress creates a warming address.
func (a *App) CreateWarmingAddress(c echo.Context) error {
	var o struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if o.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	id, err := a.core.CreateWarmingAddress(o.Email, o.Name)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: map[string]int{"id": id}})
}

// UpdateWarmingAddress updates a warming address.
func (a *App) UpdateWarmingAddress(c echo.Context) error {
	id := getID(c)
	var o struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := a.core.UpdateWarmingAddress(id, o.Email, o.Name, o.IsActive); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{true})
}

// DeleteWarmingAddress deletes a warming address.
func (a *App) DeleteWarmingAddress(c echo.Context) error {
	id := getID(c)
	if err := a.core.DeleteWarmingAddress(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{true})
}

// GetWarmingSenders returns all warming senders.
func (a *App) GetWarmingSenders(c echo.Context) error {
	out, err := a.core.GetWarmingSenders()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateWarmingSender creates a warming sender.
func (a *App) CreateWarmingSender(c echo.Context) error {
	var o struct {
		Email      string `json:"email"`
		Name       string `json:"name"`
		Brand      string `json:"brand"`
		BrandURL   string `json:"brand_url"`
		BrandColor string `json:"brand_color"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if o.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}
	if o.BrandColor == "" {
		o.BrandColor = "#F2C94C"
	}

	id, err := a.core.CreateWarmingSender(o.Email, o.Name, o.Brand, o.BrandURL, o.BrandColor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: map[string]int{"id": id}})
}

// UpdateWarmingSender updates a warming sender.
func (a *App) UpdateWarmingSender(c echo.Context) error {
	id := getID(c)
	var o struct {
		Email      string `json:"email"`
		Name       string `json:"name"`
		Brand      string `json:"brand"`
		BrandURL   string `json:"brand_url"`
		BrandColor string `json:"brand_color"`
		IsActive   bool   `json:"is_active"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := a.core.UpdateWarmingSender(id, o.Email, o.Name, o.Brand, o.BrandURL, o.BrandColor, o.IsActive); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{true})
}

// DeleteWarmingSender deletes a warming sender.
func (a *App) DeleteWarmingSender(c echo.Context) error {
	id := getID(c)
	if err := a.core.DeleteWarmingSender(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{true})
}

// GetWarmingTemplates returns all warming templates.
func (a *App) GetWarmingTemplates(c echo.Context) error {
	out, err := a.core.GetWarmingTemplates()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateWarmingTemplate creates a warming template.
func (a *App) CreateWarmingTemplate(c echo.Context) error {
	var o struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if o.Subject == "" || o.Body == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "subject and body are required")
	}

	id, err := a.core.CreateWarmingTemplate(o.Subject, o.Body)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: map[string]int{"id": id}})
}

// UpdateWarmingTemplate updates a warming template.
func (a *App) UpdateWarmingTemplate(c echo.Context) error {
	id := getID(c)
	var o struct {
		Subject  string `json:"subject"`
		Body     string `json:"body"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := a.core.UpdateWarmingTemplate(id, o.Subject, o.Body, o.IsActive); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{true})
}

// DeleteWarmingTemplate deletes a warming template.
func (a *App) DeleteWarmingTemplate(c echo.Context) error {
	id := getID(c)
	if err := a.core.DeleteWarmingTemplate(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{true})
}

// GetWarmingConfig returns the warming config.
func (a *App) GetWarmingConfig(c echo.Context) error {
	out, err := a.core.GetWarmingConfig()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: out})
}

// UpdateWarmingConfig updates the warming config.
func (a *App) UpdateWarmingConfig(c echo.Context) error {
	var o models.WarmingConfig
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := a.core.UpdateWarmingConfig(o); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{true})
}

// GetWarmingSendLog returns paginated warming send log.
func (a *App) GetWarmingSendLog(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 50
	}

	campaignID, _ := strconv.Atoi(c.QueryParam("campaign_id"))
	out, total, err := a.core.GetWarmingSendLog(offset, limit, campaignID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: struct {
		Results []models.WarmingSendLog `json:"results"`
		Total   int                     `json:"total"`
	}{out, total}})
}

// GetWarmingStats returns warming statistics.
func (a *App) GetWarmingStats(c echo.Context) error {
	count, err := a.core.GetWarmingSendsToday()
	if err != nil {
		return err
	}

	cfg, err := a.core.GetWarmingConfig()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, okResp{Data: models.WarmingStats{
		SentToday: count,
		IsActive:  cfg.IsActive,
	}})
}

// GetWarmingCampaigns returns all warming campaigns.
func (a *App) GetWarmingCampaigns(c echo.Context) error {
	out, err := a.core.GetWarmingCampaigns()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: out})
}

// CreateWarmingCampaign creates a warming campaign.
func (a *App) CreateWarmingCampaign(c echo.Context) error {
	var o struct {
		Name              string          `json:"name"`
		Brand             string          `json:"brand"`
		SenderDomains     []string        `json:"sender_domains"`
		Status            string          `json:"status"`
		SendsPerRun       int             `json:"sends_per_run"`
		RunsPerDay        int             `json:"runs_per_day"`
		ScheduleTimes     []string        `json:"schedule_times"`
		RandomDelayMin    int             `json:"random_delay_min_s"`
		RandomDelayMax    int             `json:"random_delay_max_s"`
		WarmupStartDate   *string         `json:"warmup_start_date"`
		DailyLimits       json.RawMessage `json:"daily_limits"`
		HourlyCap         int             `json:"hourly_cap"`
		BusinessHoursOnly bool            `json:"business_hours_only"`
		SenderID          *int            `json:"sender_id"`
		Messenger         string          `json:"messenger"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if o.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if o.Status == "" {
		o.Status = "draft"
	}
	if o.SendsPerRun < 1 {
		o.SendsPerRun = 3
	}
	if o.RunsPerDay < 1 {
		o.RunsPerDay = 4
	}
	if len(o.ScheduleTimes) == 0 {
		o.ScheduleTimes = []string{"10:00", "14:00", "18:00", "21:00"}
	}
	if o.RandomDelayMin < 1 {
		o.RandomDelayMin = 30
	}
	if o.RandomDelayMax < 1 {
		o.RandomDelayMax = 120
	}
	if len(o.DailyLimits) == 0 {
		o.DailyLimits = json.RawMessage("[]")
	}

	var startDate null.Time
	if o.WarmupStartDate != nil && *o.WarmupStartDate != "" {
		startDate = null.Time{Time: parseDate(*o.WarmupStartDate), Valid: true}
	}

	var senderID null.Int
	if o.SenderID != nil {
		senderID = null.NewInt(*o.SenderID, true)
	}

	id, err := a.core.CreateWarmingCampaign(models.WarmingCampaign{
		Name:              o.Name,
		Brand:             o.Brand,
		SenderDomains:     pq.StringArray(o.SenderDomains),
		Status:            o.Status,
		SendsPerRun:       o.SendsPerRun,
		RunsPerDay:        o.RunsPerDay,
		ScheduleTimes:     pq.StringArray(o.ScheduleTimes),
		RandomDelayMin:    o.RandomDelayMin,
		RandomDelayMax:    o.RandomDelayMax,
		DailyLimits:       o.DailyLimits,
		HourlyCap:         o.HourlyCap,
		BusinessHoursOnly: o.BusinessHoursOnly,
		WarmupStartDate:   startDate,
		SenderID:          senderID,
		Messenger:         o.Messenger,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: map[string]int{"id": id}})
}

// UpdateWarmingCampaign updates a warming campaign.
func (a *App) UpdateWarmingCampaign(c echo.Context) error {
	id := getID(c)
	var o struct {
		Name              string          `json:"name"`
		Brand             string          `json:"brand"`
		SenderDomains     []string        `json:"sender_domains"`
		Status            string          `json:"status"`
		SendsPerRun       int             `json:"sends_per_run"`
		RunsPerDay        int             `json:"runs_per_day"`
		ScheduleTimes     []string        `json:"schedule_times"`
		RandomDelayMin    int             `json:"random_delay_min_s"`
		RandomDelayMax    int             `json:"random_delay_max_s"`
		WarmupStartDate   *string         `json:"warmup_start_date"`
		DailyLimits       json.RawMessage `json:"daily_limits"`
		HourlyCap         int             `json:"hourly_cap"`
		BusinessHoursOnly bool            `json:"business_hours_only"`
		SenderID          *int            `json:"sender_id"`
		Messenger         string          `json:"messenger"`
	}
	if err := c.Bind(&o); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if len(o.DailyLimits) == 0 {
		o.DailyLimits = json.RawMessage("[]")
	}

	var startDate null.Time
	if o.WarmupStartDate != nil && *o.WarmupStartDate != "" {
		startDate = null.Time{Time: parseDate(*o.WarmupStartDate), Valid: true}
	}

	var senderID null.Int
	if o.SenderID != nil {
		senderID = null.NewInt(*o.SenderID, true)
	}

	if err := a.core.UpdateWarmingCampaign(id, models.WarmingCampaign{
		Name:              o.Name,
		Brand:             o.Brand,
		SenderDomains:     pq.StringArray(o.SenderDomains),
		Status:            o.Status,
		SendsPerRun:       o.SendsPerRun,
		RunsPerDay:        o.RunsPerDay,
		ScheduleTimes:     pq.StringArray(o.ScheduleTimes),
		RandomDelayMin:    o.RandomDelayMin,
		RandomDelayMax:    o.RandomDelayMax,
		DailyLimits:       o.DailyLimits,
		HourlyCap:         o.HourlyCap,
		BusinessHoursOnly: o.BusinessHoursOnly,
		WarmupStartDate:   startDate,
		SenderID:          senderID,
		Messenger:         o.Messenger,
	}); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{true})
}

// GetWarmingCampaignStats returns send statistics for a warming campaign.
func (a *App) GetWarmingCampaignStats(c echo.Context) error {
	id := getID(c)
	stats, err := a.core.GetWarmingCampaignStats(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{Data: stats})
}

// DeleteWarmingCampaign deletes a warming campaign.
func (a *App) DeleteWarmingCampaign(c echo.Context) error {
	id := getID(c)
	if err := a.core.DeleteWarmingCampaign(id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{true})
}
