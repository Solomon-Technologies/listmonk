package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/knadh/listmonk/models"
)

// resendEvent represents a Resend webhook event payload.
type resendEvent struct {
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	Data      struct {
		EmailID string   `json:"email_id"`
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Bounce  struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"bounce"`
	} `json:"data"`
}

// Resend handles Resend webhook notifications (bounces and complaints).
type Resend struct {
	signingKey []byte // base64-decoded Svix webhook signing secret (optional)
}

// NewResend creates a new Resend webhook handler.
// key is the Svix signing secret from the Resend dashboard (e.g., "whsec_...").
// If empty, signature verification is skipped.
func NewResend(key string) *Resend {
	var decoded []byte
	if key != "" {
		// Strip "whsec_" prefix if present, then base64-decode.
		raw := strings.TrimPrefix(key, "whsec_")
		if d, err := base64.StdEncoding.DecodeString(raw); err == nil {
			decoded = d
		}
	}
	return &Resend{signingKey: decoded}
}

// ProcessBounce processes a Resend webhook event and returns bounce records.
// svixID, svixTimestamp, svixSignature are from the request headers.
func (r *Resend) ProcessBounce(svixID, svixTimestamp, svixSignature string, body []byte) ([]models.Bounce, error) {
	// Verify signature if a signing key is configured.
	if len(r.signingKey) > 0 {
		if err := r.verifySignature(svixID, svixTimestamp, svixSignature, body); err != nil {
			return nil, fmt.Errorf("resend signature verification failed: %v", err)
		}
	}

	var ev resendEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		return nil, fmt.Errorf("error unmarshalling Resend event: %v", err)
	}

	// Determine bounce type from event type.
	var typ string
	switch ev.Type {
	case "email.bounced":
		typ = models.BounceTypeSoft
		if ev.Data.Bounce.Type == "Permanent" {
			typ = models.BounceTypeHard
		}
	case "email.complained":
		typ = models.BounceTypeComplaint
	case "email.delivery_delayed":
		typ = models.BounceTypeSoft
	default:
		// Ignore non-bounce events (email.sent, email.delivered, email.opened, email.clicked).
		return nil, nil
	}

	ts, _ := time.Parse(time.RFC3339, ev.CreatedAt)

	// Build one bounce per recipient.
	var bounces []models.Bounce
	for _, to := range ev.Data.To {
		bounces = append(bounces, models.Bounce{
			Email:     strings.ToLower(strings.TrimSpace(to)),
			Type:      typ,
			Source:    "resend",
			Meta:      json.RawMessage(body),
			CreatedAt: ts,
		})
	}

	return bounces, nil
}

// verifySignature verifies the Svix webhook signature used by Resend.
// Signature format: "v1,<base64_sig>" (may have multiple comma-separated signatures).
// Message to sign: "{svix-id}.{svix-timestamp}.{body}"
func (r *Resend) verifySignature(msgID, timestamp, signature string, body []byte) error {
	if msgID == "" || timestamp == "" || signature == "" {
		return errors.New("missing svix headers")
	}

	// Construct the signed content.
	msg := []byte(msgID + "." + timestamp + ".")
	msg = append(msg, body...)

	mac := hmac.New(sha256.New, r.signingKey)
	mac.Write(msg)
	expected := mac.Sum(nil)

	// The signature header can contain multiple signatures (e.g., "v1,abc v1,def").
	// Split by space and check each.
	for _, part := range strings.Split(signature, " ") {
		parts := strings.SplitN(part, ",", 2)
		if len(parts) != 2 || parts[0] != "v1" {
			continue
		}
		sig, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			continue
		}
		if hmac.Equal(expected, sig) {
			return nil
		}
	}

	return errors.New("no matching signature found")
}
