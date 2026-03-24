// Package webhooks implements an async webhook dispatcher with HMAC signing and retries.
package webhooks

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/knadh/listmonk/models"
)

// Store defines the data access interface for webhooks.
type Store interface {
	GetWebhooksByEvent(event string) (models.Webhooks, error)
	InsertWebhookLog(webhookID int, event string, payload []byte, respCode int, respBody, errMsg string, attempt int) error
}

// event represents a queued webhook event to be dispatched.
type event struct {
	Name    string
	Payload any
}

// Manager manages webhook dispatching.
type Manager struct {
	store  Store
	ch     chan event
	log    *log.Logger
	client *http.Client
}

// New creates a new webhook manager.
func New(store Store, lo *log.Logger) *Manager {
	m := &Manager{
		store: store,
		ch:    make(chan event, 1000),
		log:   lo,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	go m.worker()
	return m
}

// Dispatch queues a webhook event for async delivery to all matching endpoints.
func (m *Manager) Dispatch(eventName string, payload any) {
	select {
	case m.ch <- event{Name: eventName, Payload: payload}:
	default:
		m.log.Printf("webhook queue full, dropping event: %s", eventName)
	}
}

// worker processes queued events.
func (m *Manager) worker() {
	for ev := range m.ch {
		hooks, err := m.store.GetWebhooksByEvent(ev.Name)
		if err != nil {
			m.log.Printf("error fetching webhooks for event %s: %v", ev.Name, err)
			continue
		}

		for _, wh := range hooks {
			go m.deliver(wh, ev)
		}
	}
}

// deliver sends a webhook payload with retries and exponential backoff.
func (m *Manager) deliver(wh models.Webhook, ev event) {
	body := map[string]any{
		"event":     ev.Name,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"data":      ev.Payload,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		m.log.Printf("error marshalling webhook payload: %v", err)
		return
	}

	maxAttempts := wh.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	timeout := time.Duration(wh.TimeoutSeconds) * time.Second
	if timeout < 1*time.Second {
		timeout = 10 * time.Second
	}

	client := &http.Client{Timeout: timeout}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest("POST", wh.URL, bytes.NewReader(payload))
		if err != nil {
			m.log.Printf("error creating webhook request: %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Solomon-Listmonk-Webhook/1.0")
		req.Header.Set("X-Webhook-Event", ev.Name)

		// HMAC signature if secret is configured.
		if wh.Secret != "" {
			sig := signPayload(payload, wh.Secret)
			req.Header.Set("X-Webhook-Signature", "sha256="+sig)
		}

		resp, err := client.Do(req)
		var (
			respCode int
			respBody string
			errMsg   string
		)

		if err != nil {
			errMsg = err.Error()
		} else {
			respCode = resp.StatusCode
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
			respBody = string(b)
		}

		// Log the attempt.
		_ = m.store.InsertWebhookLog(wh.ID, ev.Name, payload, respCode, respBody, errMsg, attempt)

		// Success (2xx).
		if respCode >= 200 && respCode < 300 {
			return
		}

		// Retry with exponential backoff.
		if attempt < maxAttempts {
			backoff := time.Duration(attempt*attempt) * time.Second
			m.log.Printf("webhook %d delivery attempt %d failed (status=%d), retrying in %v", wh.ID, attempt, respCode, backoff)
			time.Sleep(backoff)
		} else {
			m.log.Printf("webhook %d delivery failed after %d attempts for event %s", wh.ID, maxAttempts, ev.Name)
		}
	}
}

// signPayload creates an HMAC-SHA256 signature of the payload.
func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// Close shuts down the webhook manager.
func (m *Manager) Close() {
	close(m.ch)
}

// Verify checks if a payload matches an HMAC-SHA256 signature.
func Verify(payload []byte, secret, signature string) bool {
	expected := signPayload(payload, secret)
	return hmac.Equal([]byte(fmt.Sprintf("sha256=%s", expected)), []byte(signature))
}
