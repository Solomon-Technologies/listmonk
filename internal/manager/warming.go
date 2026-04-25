package manager

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/knadh/listmonk/models"
)

// WarmingStore defines the data interface for warming processing.
type WarmingStore interface {
	GetWarmingConfig() (models.WarmingConfig, error)
	GetActiveWarmingAddresses() ([]models.WarmingAddress, error)
	GetActiveWarmingSenders() ([]models.WarmingSender, error)
	GetActiveWarmingTemplates() ([]models.WarmingTemplate, error)
	GetWarmingSendsToday() (int, error)
	RecordWarmingSend(senderEmail, recipientEmail string, templateID int, subject, status, errMsg string)

	// Campaign-based methods.
	GetActiveWarmingCampaigns() ([]models.WarmingCampaign, error)
	GetWarmingSendersByDomains(domains []string) ([]models.WarmingSender, error)
	GetWarmingSenderByID(id int) (models.WarmingSender, error)
	GetWarmingSendsTodayByCampaign(campaignID int) (int, error)
	GetWarmingSendsLastHourByCampaign(campaignID int) (int, error)
	SetWarmingCampaignStartDate(campaignID int) error
	RecordWarmingSendCampaign(campaignID int, senderEmail, recipientEmail string, templateID int, subject, status, errMsg string)
}

// WarmingProcessor processes scheduled warming email sends.
type WarmingProcessor struct {
	store   WarmingStore
	manager *Manager
	log     *log.Logger
	ticker  *time.Ticker
	stopCh  chan struct{}

	// Check interval (how often to poll if it is time for a run).
	interval time.Duration

	// Track last run per campaign to prevent double-runs within the same schedule window.
	lastRunAt   map[int]time.Time
	lastRunAtMu sync.Mutex
}

// WarmingProcessorConfig contains configuration.
type WarmingProcessorConfig struct {
	Interval time.Duration
}

// NewWarmingProcessor creates a new warming processor.
func NewWarmingProcessor(store WarmingStore, mgr *Manager, lo *log.Logger, cfg WarmingProcessorConfig) *WarmingProcessor {
	if cfg.Interval < 10*time.Second {
		cfg.Interval = 60 * time.Second
	}
	return &WarmingProcessor{
		store:     store,
		manager:   mgr,
		log:       lo,
		interval:  cfg.Interval,
		stopCh:    make(chan struct{}),
		lastRunAt: make(map[int]time.Time),
	}
}

// Run starts the warming processor ticker loop.
func (w *WarmingProcessor) Run() {
	w.ticker = time.NewTicker(w.interval)
	w.log.Printf("warming processor started (check interval=%v)", w.interval)

	for {
		select {
		case <-w.ticker.C:
			w.checkAndProcess()
		case <-w.stopCh:
			w.ticker.Stop()
			w.log.Println("warming processor stopped")
			return
		}
	}
}

// Close stops the warming processor.
func (w *WarmingProcessor) Close() {
	close(w.stopCh)
}

// checkAndProcess iterates over active warming campaigns and runs sends for each.
func (w *WarmingProcessor) checkAndProcess() {
	campaigns, err := w.store.GetActiveWarmingCampaigns()
	if err != nil {
		w.log.Printf("error fetching active warming campaigns: %v", err)
		return
	}

	if len(campaigns) == 0 {
		return
	}

	now := time.Now().UTC()

	// Fetch shared data once (addresses and templates are global).
	addresses, err := w.store.GetActiveWarmingAddresses()
	if err != nil || len(addresses) == 0 {
		return
	}
	templates, err := w.store.GetActiveWarmingTemplates()
	if err != nil || len(templates) == 0 {
		return
	}

	for _, camp := range campaigns {
		w.processCampaign(camp, now, addresses, templates)
	}
}

// processCampaign handles a single warming campaign's scheduled sends.
func (w *WarmingProcessor) processCampaign(camp models.WarmingCampaign, now time.Time, addresses []models.WarmingAddress, templates []models.WarmingTemplate) {
	// Per-campaign recipient subset: if recipient_ids is non-empty, restrict
	// this campaign's sends to those address IDs only. Empty = send to all
	// active addresses (default).
	if len(camp.RecipientIDs) > 0 {
		allowed := make(map[int]struct{}, len(camp.RecipientIDs))
		for _, id := range camp.RecipientIDs {
			allowed[int(id)] = struct{}{}
		}
		filtered := addresses[:0:0]
		for _, a := range addresses {
			if _, ok := allowed[a.ID]; ok {
				filtered = append(filtered, a)
			}
		}
		if len(filtered) == 0 {
			w.log.Printf("warming campaign %q: recipient_ids set but none match active addresses, skipping", camp.Name)
			return
		}
		addresses = filtered
	}

	// Business hours check: Mon-Fri, 9 AM - 6 PM ET (13:00-22:00 UTC).
	if camp.BusinessHoursOnly {
		if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
			return
		}
		if now.Hour() < 13 || now.Hour() >= 22 {
			return
		}
	}

	// Check if current time matches any of this campaign's schedule times.
	nowHM := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	matched := false

	for _, t := range camp.ScheduleTimes {
		parts := strings.Split(t, ":")
		if len(parts) != 2 {
			continue
		}

		schedH, schedM := 0, 0
		fmt.Sscanf(parts[0], "%d", &schedH)
		fmt.Sscanf(parts[1], "%d", &schedM)
		schedTime := fmt.Sprintf("%02d:%02d", schedH, schedM)

		if nowHM == schedTime || (now.Minute() == schedM+1 && now.Hour() == schedH) {
			matched = true
			break
		}
	}

	if !matched {
		return
	}

	// Prevent double-runs within the same 5-minute window per campaign.
	w.lastRunAtMu.Lock()
	if last, ok := w.lastRunAt[camp.ID]; ok && now.Sub(last) < 5*time.Minute {
		w.lastRunAtMu.Unlock()
		return
	}
	w.lastRunAt[camp.ID] = now
	w.lastRunAtMu.Unlock()

	// Auto-set warmup start date on first active run.
	if err := w.store.SetWarmingCampaignStartDate(camp.ID); err != nil {
		w.log.Printf("error setting warmup start date for campaign %d: %v", camp.ID, err)
	}

	// Determine the effective daily cap using progressive ramp.
	dailyCap := w.getEffectiveDailyCap(camp, now)

	// Check daily cap for this campaign.
	todayCount, err := w.store.GetWarmingSendsTodayByCampaign(camp.ID)
	if err != nil {
		w.log.Printf("error checking warming daily count for campaign %d: %v", camp.ID, err)
		return
	}
	if dailyCap > 0 && todayCount >= dailyCap {
		return
	}

	// Check hourly cap.
	if camp.HourlyCap > 0 {
		hourCount, err := w.store.GetWarmingSendsLastHourByCampaign(camp.ID)
		if err != nil {
			w.log.Printf("error checking warming hourly count for campaign %d: %v", camp.ID, err)
			return
		}
		if hourCount >= camp.HourlyCap {
			return
		}
	}

	// Get sender(s) for this campaign.
	var senders []models.WarmingSender
	if camp.SenderID.Valid {
		s, err := w.store.GetWarmingSenderByID(camp.SenderID.Int)
		if err != nil || s.ID == 0 {
			w.log.Printf("warming campaign %q: sender_id %d not found or inactive, skipping", camp.Name, camp.SenderID.Int)
			return
		}
		senders = []models.WarmingSender{s}
	} else {
		var err error
		senders, err = w.store.GetWarmingSendersByDomains(camp.SenderDomains)
		if err != nil || len(senders) == 0 {
			return
		}
	}

	// Calculate how many to send this run (don't exceed daily or hourly cap).
	toSend := camp.SendsPerRun
	if dailyCap > 0 {
		remaining := dailyCap - todayCount
		if remaining < toSend {
			toSend = remaining
		}
	}
	if camp.HourlyCap > 0 {
		hourCount, _ := w.store.GetWarmingSendsLastHourByCampaign(camp.ID)
		hourRemaining := camp.HourlyCap - hourCount
		if hourRemaining < toSend {
			toSend = hourRemaining
		}
	}
	if toSend <= 0 {
		return
	}

	w.log.Printf("warming campaign %q: sending %d emails (day cap=%d, today=%d, hourly_cap=%d)",
		camp.Name, toSend, dailyCap, todayCount, camp.HourlyCap)

	dateStr := now.Format("January 2, 2006")

	for i := 0; i < toSend; i++ {
		sender := senders[i%len(senders)]
		addr := addresses[rand.Intn(len(addresses))]
		tpl := templates[rand.Intn(len(templates))]

		// Replace placeholders.
		subject := strings.ReplaceAll(tpl.Subject, "{{name}}", addr.Name)
		subject = strings.ReplaceAll(subject, "{{date}}", dateStr)
		body := strings.ReplaceAll(tpl.Body, "{{name}}", addr.Name)
		body = strings.ReplaceAll(body, "{{date}}", dateStr)

		// Build plain-text HTML (looks like a human email).
		htmlBody := fmt.Sprintf(
			`<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; font-size: 15px; color: #0f172a; line-height: 1.6; max-width: 600px;">
<p>%s</p>
<p style="margin-top: 24px;">Best,<br><strong>%s</strong><br><a href="%s" style="color: %s; text-decoration: none;">%s</a></p>
</div>`,
			body, sender.Name, sender.BrandURL, sender.BrandColor, sender.Brand)

		fromEmail := fmt.Sprintf("%s <%s>", sender.Name, sender.Email)

		msg := models.Message{
			From:        fromEmail,
			To:          []string{addr.Email},
			Subject:     subject,
			ContentType: "html",
			Body:        []byte(htmlBody),
			Messenger: func() string {
			if camp.Messenger != "" {
				return camp.Messenger
			}
			return "email"
		}(),
		}

		if err := w.manager.PushMessage(msg); err != nil {
			w.log.Printf("warming send error [%s] (%s -> %s): %v", camp.Name, sender.Email, addr.Email, err)
			w.store.RecordWarmingSendCampaign(camp.ID, sender.Email, addr.Email, tpl.ID, subject, "failed", err.Error())
		} else {
			w.store.RecordWarmingSendCampaign(camp.ID, sender.Email, addr.Email, tpl.ID, subject, "sent", "")
		}

		// Random delay between sends.
		if i < toSend-1 && camp.RandomDelayMax > camp.RandomDelayMin {
			delay := camp.RandomDelayMin + rand.Intn(camp.RandomDelayMax-camp.RandomDelayMin)
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}

	w.log.Printf("warming campaign %q: run complete (%d emails)", camp.Name, toSend)
}

// getEffectiveDailyCap returns the daily send limit for this campaign, accounting
// for the progressive ramp schedule. If daily_limits is set and warmup_start_date
// exists, it uses the day-based schedule. Otherwise falls back to sends_per_run * runs_per_day.
func (w *WarmingProcessor) getEffectiveDailyCap(camp models.WarmingCampaign, now time.Time) int {
	defaultCap := camp.SendsPerRun * camp.RunsPerDay

	// If no daily limits configured, use the flat cap.
	if len(camp.DailyLimits) == 0 || string(camp.DailyLimits) == "[]" || string(camp.DailyLimits) == "null" {
		return defaultCap
	}

	// If no start date, can't calculate day number.
	if !camp.WarmupStartDate.Valid {
		return defaultCap
	}

	var limits []models.DailyLimit
	if err := json.Unmarshal(camp.DailyLimits, &limits); err != nil || len(limits) == 0 {
		return defaultCap
	}

	// Calculate which day of the warmup we're on (1-indexed).
	startDate := camp.WarmupStartDate.Time
	dayNum := int(now.Sub(startDate).Hours()/24) + 1
	if dayNum < 1 {
		dayNum = 1
	}

	// Find the matching limit. Use the last entry if we've exceeded the schedule.
	effectiveCap := defaultCap
	for _, dl := range limits {
		if dayNum <= dl.Day {
			effectiveCap = dl.Max
			break
		}
		// If we're past this entry, use it as the floor and keep looking.
		effectiveCap = dl.Max
	}

	return effectiveCap
}
