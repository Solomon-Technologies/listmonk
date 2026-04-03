package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/textproto"
	"strings"
	txttpl "text/template"
	"time"

	"github.com/knadh/listmonk/models"
)

// DripStore defines the data interface for drip campaign processing.
type DripStore interface {
	GetPendingDripSends(limit int) ([]models.PendingDripSend, error)
	AdvanceDripEnrollment(enrollmentID int64, currentStepID, dripCampaignID int) error
	RecordDripSend(dripCampaignID, stepID, subscriberID int, status, errMsg string)
	GetDripSendsToday(dripCampaignID int) (int, error)
	UpdateDripStepSent(stepID int) error
	UpdateDripCampaignEntered(campaignID int) error
	UpdateDripCampaignCompleted(campaignID int) error
}

// DripMessage represents a drip email being rendered for a specific subscriber.
// It serves as the "dot" context for Go templates.
type DripMessage struct {
	DripCampaign *dripCampaignInfo
	DripStep     *dripStepInfo
	Subscriber   models.Subscriber

	from    string
	to      string
	subject string
	body    []byte
	altBody []byte

	tpl        *template.Template
	subjectTpl *txttpl.Template
	altBodyTpl *template.Template
}

type dripCampaignInfo struct {
	ID        int
	UUID      string
	Name      string
	FromEmail string
}

type dripStepInfo struct {
	ID          int
	UUID        string
	Subject     string
	Body        string
	AltBody     string
	ContentType string
	FromEmail   string
}

// compileTemplate compiles the drip step body into a renderable template using
// the same base-template + content-template pattern as campaigns.
func (dm *DripMessage) compileTemplate(f template.FuncMap, baseTemplateBody string) error {
	// Compile subject if it contains template expressions.
	if strings.Contains(dm.DripStep.Subject, "{{") {
		subj := models.ApplyTplFuncReplacements(dm.DripStep.Subject)
		var txtFuncs map[string]any = f
		subjTpl, err := txttpl.New(ContentTpl).Funcs(txtFuncs).Parse(subj)
		if err != nil {
			return fmt.Errorf("error compiling drip subject: %v", err)
		}
		dm.subjectTpl = subjTpl
	}

	// Base template body (from the step's template_id, or a passthrough wrapper).
	body := baseTemplateBody
	if body == "" {
		body = `{{ template "content" . }}`
	}
	body = models.ApplyTplFuncReplacements(body)

	baseTpl, err := template.New(BaseTPL).Funcs(f).Parse(body)
	if err != nil {
		return fmt.Errorf("error compiling drip base template: %v", err)
	}

	// Content body (the actual step message).
	contentBody := dm.DripStep.Body
	if dm.DripStep.ContentType == models.CampaignContentTypeMarkdown {
		htmlBytes, err := models.ConvertMarkdown([]byte(contentBody))
		if err != nil {
			return fmt.Errorf("error converting markdown: %v", err)
		}
		contentBody = string(htmlBytes)
	}
	contentBody = models.ApplyTplFuncReplacements(contentBody)

	msgTpl, err := template.New(ContentTpl).Funcs(f).Parse(contentBody)
	if err != nil {
		return fmt.Errorf("error compiling drip content: %v", err)
	}

	out, err := baseTpl.AddParseTree(ContentTpl, msgTpl.Tree)
	if err != nil {
		return fmt.Errorf("error inserting drip content template: %v", err)
	}
	dm.tpl = out

	// Alt body (plain text alternative).
	if dm.DripStep.AltBody != "" && strings.Contains(dm.DripStep.AltBody, "{{") {
		ab := models.ApplyTplFuncReplacements(dm.DripStep.AltBody)
		abTpl, err := template.New(ContentTpl).Funcs(f).Parse(ab)
		if err != nil {
			return fmt.Errorf("error compiling drip alt body: %v", err)
		}
		dm.altBodyTpl = abTpl
	}

	return nil
}

// render executes the compiled templates with the DripMessage as the dot context.
func (dm *DripMessage) render() error {
	// Render subject.
	if dm.subjectTpl != nil {
		var buf bytes.Buffer
		if err := dm.subjectTpl.Execute(&buf, dm); err != nil {
			return fmt.Errorf("error rendering drip subject: %v", err)
		}
		dm.subject = buf.String()
	} else {
		dm.subject = dm.DripStep.Subject
	}

	// Render body.
	var buf bytes.Buffer
	if err := dm.tpl.ExecuteTemplate(&buf, BaseTPL, dm); err != nil {
		return fmt.Errorf("error rendering drip body: %v", err)
	}
	dm.body = buf.Bytes()

	// Render alt body.
	if dm.altBodyTpl != nil {
		var abuf bytes.Buffer
		if err := dm.altBodyTpl.Execute(&abuf, dm); err != nil {
			return fmt.Errorf("error rendering drip alt body: %v", err)
		}
		dm.altBody = abuf.Bytes()
	} else if dm.DripStep.AltBody != "" {
		dm.altBody = []byte(dm.DripStep.AltBody)
	}

	return nil
}

// DripProcessor processes pending drip campaign sends.
type DripProcessor struct {
	store   DripStore
	manager *Manager
	log     *log.Logger
	ticker  *time.Ticker
	stopCh  chan struct{}

	// Callbacks.
	fnNotify  func(subject, tplName string, data any) error
	fnWebhook func(event string, payload any)

	// Config.
	batchSize int
	interval  time.Duration
}

// DripConfig contains configuration for the drip processor.
type DripConfig struct {
	BatchSize int
	Interval  time.Duration
	FnNotify  func(subject, tplName string, data any) error
	FnWebhook func(event string, payload any)
}

// NewDripProcessor creates a new drip processor.
func NewDripProcessor(store DripStore, mgr *Manager, lo *log.Logger, cfg DripConfig) *DripProcessor {
	if cfg.BatchSize < 1 {
		cfg.BatchSize = 100
	}
	if cfg.Interval < 1*time.Second {
		cfg.Interval = 30 * time.Second
	}

	return &DripProcessor{
		store:     store,
		manager:   mgr,
		log:       lo,
		fnNotify:  cfg.FnNotify,
		fnWebhook: cfg.FnWebhook,
		batchSize: cfg.BatchSize,
		interval:  cfg.Interval,
		stopCh:    make(chan struct{}),
	}
}

// Run starts the drip processor ticker loop.
func (d *DripProcessor) Run() {
	d.ticker = time.NewTicker(d.interval)
	d.log.Printf("drip processor started (interval=%v, batch=%d)", d.interval, d.batchSize)

	for {
		select {
		case <-d.ticker.C:
			d.process()
		case <-d.stopCh:
			d.ticker.Stop()
			d.log.Println("drip processor stopped")
			return
		}
	}
}

// Close stops the drip processor.
func (d *DripProcessor) Close() {
	close(d.stopCh)
}

// process fetches pending drip sends and dispatches them through the manager.
func (d *DripProcessor) process() {
	// Business hours check: only send Mon-Fri, 9 AM - 6 PM ET (13:00-22:00 UTC).
	now := time.Now().UTC()
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return
	}
	if now.Hour() < 13 || now.Hour() >= 22 {
		return
	}

	sends, err := d.store.GetPendingDripSends(d.batchSize)
	if err != nil {
		d.log.Printf("error fetching pending drip sends: %v", err)
		return
	}

	if len(sends) == 0 {
		return
	}

	d.log.Printf("processing %d pending drip sends", len(sends))

	// Track per-campaign daily send counts to avoid querying DB for every message.
	dailyCounts := make(map[int]int)

	for _, s := range sends {
		d.sendDripMessage(s, dailyCounts)
	}
}

// sendDripMessage compiles templates, renders subscriber-specific content with
// tracking links/pixels, and sends a single drip message through the manager.
func (d *DripProcessor) sendDripMessage(s models.PendingDripSend, dailyCounts map[int]int) {
	// 1. Check daily send limit.
	if s.MaxSendPerDay > 0 {
		count, ok := dailyCounts[s.DripCampaignID]
		if !ok {
			var err error
			count, err = d.store.GetDripSendsToday(s.DripCampaignID)
			if err != nil {
				d.log.Printf("error checking daily send count for drip %d: %v", s.DripCampaignID, err)
				return
			}
			dailyCounts[s.DripCampaignID] = count
		}

		if count >= s.MaxSendPerDay {
			return // At daily limit, skip silently.
		}
	}

	// 2. Determine from address.
	fromEmail := s.StepFromEmail
	if fromEmail == "" {
		fromEmail = s.CampaignFromEmail
	}
	if fromEmail == "" {
		fromEmail = d.manager.cfg.FromEmail
	}

	// 3. Build subscriber model.
	var attribs models.JSON
	if len(s.SubscriberAttribs) > 0 {
		_ = json.Unmarshal(s.SubscriberAttribs, &attribs)
	}
	sub := models.Subscriber{
		Base:    models.Base{ID: s.SubscriberID},
		UUID:    s.SubscriberUUID,
		Email:   s.SubscriberEmail,
		Name:    s.SubscriberName,
		Status:  s.SubscriberStatus,
		Attribs: attribs,
	}

	// 4. Build DripMessage (the template dot context).
	dm := &DripMessage{
		DripCampaign: &dripCampaignInfo{
			ID: s.DripCampaignID, UUID: s.CampaignUUID,
			Name: s.CampaignName, FromEmail: s.CampaignFromEmail,
		},
		DripStep: &dripStepInfo{
			ID: s.CurrentStepID, UUID: s.StepUUID,
			Subject: s.Subject, Body: s.Body,
			AltBody: s.AltBody, ContentType: s.ContentType,
			FromEmail: s.StepFromEmail,
		},
		Subscriber: sub,
		from:       fromEmail,
		to:         s.SubscriberEmail,
	}

	// 5. Compile and render template with subscriber variables + tracking.
	funcMap := d.manager.DripTemplateFuncs(s.CampaignUUID, s.StepUUID)
	if err := dm.compileTemplate(funcMap, s.TemplateBody); err != nil {
		d.log.Printf("error compiling drip template (step %d, sub %d): %v", s.CurrentStepID, s.SubscriberID, err)
		d.store.RecordDripSend(s.DripCampaignID, s.CurrentStepID, s.SubscriberID, "failed", err.Error())
		return
	}
	if err := dm.render(); err != nil {
		d.log.Printf("error rendering drip message (step %d, sub %s): %v", s.CurrentStepID, s.SubscriberEmail, err)
		d.store.RecordDripSend(s.DripCampaignID, s.CurrentStepID, s.SubscriberID, "failed", err.Error())
		return
	}

	// 6. Build unsubscribe URL.
	unsubURL := fmt.Sprintf(d.manager.cfg.UnsubURL, s.CampaignUUID, s.SubscriberUUID)

	// 7. Build the final message.
	msg := models.Message{
		From:        fromEmail,
		To:          []string{s.SubscriberEmail},
		Subject:     dm.subject,
		ContentType: s.ContentType,
		Body:        dm.body,
		AltBody:     dm.altBody,
		Messenger:   s.Messenger,
		Subscriber:  sub,
	}
	if msg.Messenger == "" {
		msg.Messenger = "email"
	}

	// 8. Set e-mail headers (tracking, unsubscribe, custom).
	h := textproto.MIMEHeader{}
	h.Set(models.EmailHeaderCampaignUUID, s.CampaignUUID)
	h.Set(models.EmailHeaderSubscriberUUID, s.SubscriberUUID)
	h.Set("X-Listmonk-Drip-Campaign", s.CampaignName)

	if d.manager.cfg.UnsubHeader {
		h.Set("List-Unsubscribe-Post", "List-Unsubscribe=One-Click")
		h.Set("List-Unsubscribe", "<"+unsubURL+">")
	}
	msg.Headers = h

	// 9. Push through the manager's message queue (respects SMTP rate limits).
	if err := d.manager.PushMessage(msg); err != nil {
		d.log.Printf("error sending drip message to %s: %v", s.SubscriberEmail, err)
		d.store.RecordDripSend(s.DripCampaignID, s.CurrentStepID, s.SubscriberID, "failed", err.Error())

		// Webhook: step failed.
		if d.fnWebhook != nil {
			d.fnWebhook("drip.step.failed", map[string]any{
				"drip_campaign_id": s.DripCampaignID,
				"campaign_name":    s.CampaignName,
				"step_id":          s.CurrentStepID,
				"subscriber_id":    s.SubscriberID,
				"subscriber_email": s.SubscriberEmail,
				"error":            err.Error(),
			})
		}
		return
	}

	// 10. Record success + update counters.
	d.store.RecordDripSend(s.DripCampaignID, s.CurrentStepID, s.SubscriberID, "sent", "")
	d.store.UpdateDripStepSent(s.CurrentStepID)

	// Track daily count in memory.
	dailyCounts[s.DripCampaignID]++

	// Webhook: step sent.
	if d.fnWebhook != nil {
		d.fnWebhook("drip.step.sent", map[string]any{
			"drip_campaign_id": s.DripCampaignID,
			"campaign_name":    s.CampaignName,
			"step_id":          s.CurrentStepID,
			"subscriber_id":    s.SubscriberID,
			"subscriber_email": s.SubscriberEmail,
		})
	}

	// 11. Advance to next step or mark complete.
	if err := d.store.AdvanceDripEnrollment(s.EnrollmentID, s.CurrentStepID, s.DripCampaignID); err != nil {
		d.log.Printf("error advancing drip enrollment %d: %v", s.EnrollmentID, err)
	}
}
