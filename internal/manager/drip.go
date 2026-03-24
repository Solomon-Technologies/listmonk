package manager

import (
	"log"
	"time"

	"github.com/knadh/listmonk/models"
)

// DripStore defines the data interface for drip campaign processing.
type DripStore interface {
	GetPendingDripSends(limit int) ([]models.PendingDripSend, error)
	AdvanceDripEnrollment(enrollmentID int64, currentStepID, dripCampaignID int) error
	RecordDripSend(dripCampaignID, stepID, subscriberID int, status, errMsg string)
}

// DripProcessor processes pending drip campaign sends.
type DripProcessor struct {
	store    DripStore
	manager  *Manager
	log      *log.Logger
	ticker   *time.Ticker
	stopCh   chan struct{}

	// Config.
	batchSize int
	interval  time.Duration
}

// DripConfig contains configuration for the drip processor.
type DripConfig struct {
	BatchSize int
	Interval  time.Duration
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
	sends, err := d.store.GetPendingDripSends(d.batchSize)
	if err != nil {
		d.log.Printf("error fetching pending drip sends: %v", err)
		return
	}

	if len(sends) == 0 {
		return
	}

	d.log.Printf("processing %d pending drip sends", len(sends))

	for _, s := range sends {
		d.sendDripMessage(s)
	}
}

// sendDripMessage renders and sends a single drip message.
func (d *DripProcessor) sendDripMessage(s models.PendingDripSend) {
	// Determine from address.
	fromEmail := s.StepFromEmail
	if fromEmail == "" {
		fromEmail = s.CampaignFromEmail
	}
	if fromEmail == "" {
		fromEmail = d.manager.cfg.FromEmail
	}

	// Build the message.
	msg := models.Message{
		From:        fromEmail,
		To:          []string{s.SubscriberEmail},
		Subject:     s.Subject,
		ContentType: s.ContentType,
		Body:        []byte(s.Body),
		AltBody:     []byte(s.AltBody),
		Messenger:   s.Messenger,
		Subscriber: models.Subscriber{
			Base:   models.Base{ID: s.SubscriberID},
			UUID:   s.SubscriberUUID,
			Email:  s.SubscriberEmail,
			Name:   s.SubscriberName,
			Status: s.SubscriberStatus,
		},
	}

	if msg.Messenger == "" {
		msg.Messenger = "email"
	}

	// Push through the manager's message queue (respects rate limits, SMTP config).
	if err := d.manager.PushMessage(msg); err != nil {
		d.log.Printf("error sending drip message to %s: %v", s.SubscriberEmail, err)
		d.store.RecordDripSend(s.DripCampaignID, s.CurrentStepID, s.SubscriberID, "failed", err.Error())
		return
	}

	// Record success.
	d.store.RecordDripSend(s.DripCampaignID, s.CurrentStepID, s.SubscriberID, "sent", "")

	// Advance to next step or mark complete.
	if err := d.store.AdvanceDripEnrollment(s.EnrollmentID, s.CurrentStepID, s.DripCampaignID); err != nil {
		d.log.Printf("error advancing drip enrollment %d: %v", s.EnrollmentID, err)
	}
}
