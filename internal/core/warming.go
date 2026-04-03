package core

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

// GetWarmingAddresses returns all warming addresses.
func (c *Core) GetWarmingAddresses() ([]models.WarmingAddress, error) {
	var out []models.WarmingAddress
	if err := c.q.GetWarmingAddresses.Select(&out); err != nil {
		c.log.Printf("error fetching warming addresses: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming addresses", "error", pqErrMsg(err)))
	}
	return out, nil
}

// CreateWarmingAddress creates a new warming address.
func (c *Core) CreateWarmingAddress(email, name string) (int, error) {
	var id int
	if err := c.q.CreateWarmingAddress.Get(&id, email, name); err != nil {
		c.log.Printf("error creating warming address: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "warming address", "error", pqErrMsg(err)))
	}
	return id, nil
}

// UpdateWarmingAddress updates a warming address.
func (c *Core) UpdateWarmingAddress(id int, email, name string, isActive bool) error {
	if _, err := c.q.UpdateWarmingAddress.Exec(id, email, name, isActive); err != nil {
		c.log.Printf("error updating warming address: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "warming address", "error", pqErrMsg(err)))
	}
	return nil
}

// DeleteWarmingAddress deletes a warming address.
func (c *Core) DeleteWarmingAddress(id int) error {
	if _, err := c.q.DeleteWarmingAddress.Exec(id); err != nil {
		c.log.Printf("error deleting warming address: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "warming address", "error", pqErrMsg(err)))
	}
	return nil
}

// GetActiveWarmingAddresses returns only active warming addresses.
func (c *Core) GetActiveWarmingAddresses() ([]models.WarmingAddress, error) {
	var out []models.WarmingAddress
	if err := c.q.GetActiveWarmingAddresses.Select(&out); err != nil {
		c.log.Printf("error fetching active warming addresses: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming addresses", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetWarmingSenders returns all warming senders.
func (c *Core) GetWarmingSenders() ([]models.WarmingSender, error) {
	var out []models.WarmingSender
	if err := c.q.GetWarmingSenders.Select(&out); err != nil {
		c.log.Printf("error fetching warming senders: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming senders", "error", pqErrMsg(err)))
	}
	return out, nil
}

// CreateWarmingSender creates a new warming sender.
func (c *Core) CreateWarmingSender(email, name, brand, brandURL, brandColor string) (int, error) {
	var id int
	if err := c.q.CreateWarmingSender.Get(&id, email, name, brand, brandURL, brandColor); err != nil {
		c.log.Printf("error creating warming sender: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "warming sender", "error", pqErrMsg(err)))
	}
	return id, nil
}

// UpdateWarmingSender updates a warming sender.
func (c *Core) UpdateWarmingSender(id int, email, name, brand, brandURL, brandColor string, isActive bool) error {
	if _, err := c.q.UpdateWarmingSender.Exec(id, email, name, brand, brandURL, brandColor, isActive); err != nil {
		c.log.Printf("error updating warming sender: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "warming sender", "error", pqErrMsg(err)))
	}
	return nil
}

// DeleteWarmingSender deletes a warming sender.
func (c *Core) DeleteWarmingSender(id int) error {
	if _, err := c.q.DeleteWarmingSender.Exec(id); err != nil {
		c.log.Printf("error deleting warming sender: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "warming sender", "error", pqErrMsg(err)))
	}
	return nil
}

// GetActiveWarmingSenders returns only active warming senders.
func (c *Core) GetActiveWarmingSenders() ([]models.WarmingSender, error) {
	var out []models.WarmingSender
	if err := c.q.GetActiveWarmingSenders.Select(&out); err != nil {
		c.log.Printf("error fetching active warming senders: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming senders", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetWarmingTemplates returns all warming templates.
func (c *Core) GetWarmingTemplates() ([]models.WarmingTemplate, error) {
	var out []models.WarmingTemplate
	if err := c.q.GetWarmingTemplates.Select(&out); err != nil {
		c.log.Printf("error fetching warming templates: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming templates", "error", pqErrMsg(err)))
	}
	return out, nil
}

// CreateWarmingTemplate creates a new warming template.
func (c *Core) CreateWarmingTemplate(subject, body string) (int, error) {
	var id int
	if err := c.q.CreateWarmingTemplate.Get(&id, subject, body); err != nil {
		c.log.Printf("error creating warming template: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "warming template", "error", pqErrMsg(err)))
	}
	return id, nil
}

// UpdateWarmingTemplate updates a warming template.
func (c *Core) UpdateWarmingTemplate(id int, subject, body string, isActive bool) error {
	if _, err := c.q.UpdateWarmingTemplate.Exec(id, subject, body, isActive); err != nil {
		c.log.Printf("error updating warming template: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "warming template", "error", pqErrMsg(err)))
	}
	return nil
}

// DeleteWarmingTemplate deletes a warming template.
func (c *Core) DeleteWarmingTemplate(id int) error {
	if _, err := c.q.DeleteWarmingTemplate.Exec(id); err != nil {
		c.log.Printf("error deleting warming template: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "warming template", "error", pqErrMsg(err)))
	}
	return nil
}

// GetActiveWarmingTemplates returns only active warming templates.
func (c *Core) GetActiveWarmingTemplates() ([]models.WarmingTemplate, error) {
	var out []models.WarmingTemplate
	if err := c.q.GetActiveWarmingTemplates.Select(&out); err != nil {
		c.log.Printf("error fetching active warming templates: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming templates", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetWarmingConfig returns the singleton warming config.
func (c *Core) GetWarmingConfig() (models.WarmingConfig, error) {
	var out models.WarmingConfig
	if err := c.q.GetWarmingConfig.Get(&out); err != nil {
		c.log.Printf("error fetching warming config: %v", err)
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming config", "error", pqErrMsg(err)))
	}
	return out, nil
}

// UpdateWarmingConfig updates the singleton warming config.
func (c *Core) UpdateWarmingConfig(cfg models.WarmingConfig) error {
	if _, err := c.q.UpdateWarmingConfig.Exec(cfg.SendsPerRun, cfg.RunsPerDay, cfg.ScheduleTimes,
		cfg.RandomDelayMin, cfg.RandomDelayMax, cfg.IsActive); err != nil {
		c.log.Printf("error updating warming config: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "warming config", "error", pqErrMsg(err)))
	}
	return nil
}

// GetWarmingSendsToday returns the count of warming sends today.
func (c *Core) GetWarmingSendsToday() (int, error) {
	var count int
	if err := c.q.GetWarmingSendsToday.Get(&count); err != nil {
		c.log.Printf("error counting warming sends today: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming stats", "error", pqErrMsg(err)))
	}
	return count, nil
}

// RecordWarmingSend inserts a warming send log entry (legacy, no campaign).
func (c *Core) RecordWarmingSend(senderEmail, recipientEmail string, templateID int, subject, status, errMsg string) {
	if _, err := c.q.InsertWarmingSendLog.Exec(senderEmail, recipientEmail, templateID, subject, status, errMsg); err != nil {
		c.log.Printf("error recording warming send: %v", err)
	}
}

// RecordWarmingSendCampaign inserts a warming send log entry with campaign ID.
func (c *Core) RecordWarmingSendCampaign(campaignID int, senderEmail, recipientEmail string, templateID int, subject, status, errMsg string) {
	if _, err := c.q.InsertWarmingSendLogCampaign.Exec(campaignID, senderEmail, recipientEmail, templateID, subject, status, errMsg); err != nil {
		c.log.Printf("error recording warming send: %v", err)
	}
}

// GetWarmingCampaigns returns all warming campaigns.
func (c *Core) GetWarmingCampaigns() ([]models.WarmingCampaign, error) {
	var out []models.WarmingCampaign
	if err := c.q.GetWarmingCampaigns.Select(&out); err != nil {
		c.log.Printf("error fetching warming campaigns: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming campaigns", "error", pqErrMsg(err)))
	}
	return out, nil
}

// CreateWarmingCampaign creates a new warming campaign.
func (c *Core) CreateWarmingCampaign(o models.WarmingCampaign) (int, error) {
	var id int
	if err := c.q.CreateWarmingCampaign.Get(&id,
		o.Name, o.Brand, o.SenderDomains, o.Status,
		o.SendsPerRun, o.RunsPerDay, o.ScheduleTimes,
		o.RandomDelayMin, o.RandomDelayMax,
		o.WarmupStartDate, o.DailyLimits, o.HourlyCap, o.BusinessHoursOnly, o.SenderID, o.Messenger); err != nil {
		c.log.Printf("error creating warming campaign: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "warming campaign", "error", pqErrMsg(err)))
	}
	return id, nil
}

// UpdateWarmingCampaign updates a warming campaign.
func (c *Core) UpdateWarmingCampaign(id int, o models.WarmingCampaign) error {
	if _, err := c.q.UpdateWarmingCampaign.Exec(id,
		o.Name, o.Brand, o.SenderDomains, o.Status,
		o.SendsPerRun, o.RunsPerDay, o.ScheduleTimes,
		o.RandomDelayMin, o.RandomDelayMax,
		o.WarmupStartDate, o.DailyLimits, o.HourlyCap, o.BusinessHoursOnly, o.SenderID, o.Messenger); err != nil {
		c.log.Printf("error updating warming campaign: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "warming campaign", "error", pqErrMsg(err)))
	}
	return nil
}

// DeleteWarmingCampaign deletes a warming campaign.
func (c *Core) DeleteWarmingCampaign(id int) error {
	if _, err := c.q.DeleteWarmingCampaign.Exec(id); err != nil {
		c.log.Printf("error deleting warming campaign: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "warming campaign", "error", pqErrMsg(err)))
	}
	return nil
}

// GetActiveWarmingCampaigns returns active warming campaigns.
func (c *Core) GetActiveWarmingCampaigns() ([]models.WarmingCampaign, error) {
	var out []models.WarmingCampaign
	if err := c.q.GetActiveWarmingCampaigns.Select(&out); err != nil {
		c.log.Printf("error fetching active warming campaigns: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming campaigns", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetWarmingSendersByDomains returns active senders matching the given domains.
func (c *Core) GetWarmingSendersByDomains(domains []string) ([]models.WarmingSender, error) {
	var out []models.WarmingSender
	if err := c.q.GetWarmingSendersByDomains.Select(&out, pq.StringArray(domains)); err != nil {
		c.log.Printf("error fetching warming senders by domains: %v", err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming senders", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetWarmingSendsTodayByCampaign returns today's send count for a specific campaign.
func (c *Core) GetWarmingSendsTodayByCampaign(campaignID int) (int, error) {
	var count int
	if err := c.q.GetWarmingSendsTodayByCampaign.Get(&count, campaignID); err != nil {
		c.log.Printf("error counting warming sends for campaign %d: %v", campaignID, err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming stats", "error", pqErrMsg(err)))
	}
	return count, nil
}

// GetWarmingSendsLastHourByCampaign returns the count of sends in the last hour for a campaign.
func (c *Core) GetWarmingSendsLastHourByCampaign(campaignID int) (int, error) {
	var count int
	if err := c.q.GetWarmingSendsLastHourByCampaign.Get(&count, campaignID); err != nil {
		c.log.Printf("error counting warming sends last hour for campaign %d: %v", campaignID, err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming stats", "error", pqErrMsg(err)))
	}
	return count, nil
}

// SetWarmingCampaignStartDate sets the warmup_start_date to today if not already set.
func (c *Core) SetWarmingCampaignStartDate(campaignID int) error {
	if _, err := c.q.SetWarmingCampaignStartDate.Exec(campaignID); err != nil {
		c.log.Printf("error setting warming campaign start date %d: %v", campaignID, err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "warming campaign", "error", pqErrMsg(err)))
	}
	return nil
}

// GetWarmingSenderByID returns a single active warming sender by ID.
func (c *Core) GetWarmingSenderByID(id int) (models.WarmingSender, error) {
	var out models.WarmingSender
	if err := c.q.GetWarmingSenderByID.Get(&out, id); err != nil {
		c.log.Printf("error fetching warming sender %d: %v", id, err)
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming sender", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetWarmingCampaignStats returns send statistics for a single warming campaign.
func (c *Core) GetWarmingCampaignStats(campaignID int) (models.WarmingCampaignStats, error) {
	var out models.WarmingCampaignStats
	if err := c.q.GetWarmingCampaignStatsByID.Get(&out, campaignID); err != nil {
		c.log.Printf("error fetching warming campaign stats for %d: %v", campaignID, err)
		return out, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming campaign stats", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetWarmingSendLog returns paginated warming send log, optionally filtered by campaign.
func (c *Core) GetWarmingSendLog(offset, limit, campaignID int) ([]models.WarmingSendLog, int, error) {
	if limit < 1 || limit > 500 {
		limit = 50
	}

	var out []models.WarmingSendLog

	if campaignID > 0 {
		if err := c.db.Select(&out, sqlx.Rebind(sqlx.DOLLAR, c.q.GetWarmingSendLogByCampaign), limit, offset, campaignID); err != nil {
			c.log.Printf("error fetching warming send log: %v", err)
			return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
				c.i18n.Ts("globals.messages.errorFetching", "name", "warming send log", "error", pqErrMsg(err)))
		}

		var total int
		if err := c.q.GetWarmingSendLogCountByCampaign.Get(&total, campaignID); err != nil {
			c.log.Printf("error counting warming send log: %v", err)
			return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
				c.i18n.Ts("globals.messages.errorFetching", "name", "warming send log count", "error", pqErrMsg(err)))
		}
		return out, total, nil
	}

	if err := c.db.Select(&out, sqlx.Rebind(sqlx.DOLLAR, c.q.GetWarmingSendLog), limit, offset); err != nil {
		c.log.Printf("error fetching warming send log: %v", err)
		return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming send log", "error", pqErrMsg(err)))
	}

	var total int
	if err := c.q.GetWarmingSendLogCount.Get(&total); err != nil {
		c.log.Printf("error counting warming send log: %v", err)
		return nil, 0, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "warming send log count", "error", pqErrMsg(err)))
	}

	return out, total, nil
}
