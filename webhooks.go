package mailtrap

import (
	"context"
	"fmt"
	"net/http"
)

// WebhooksService manages account webhooks.
type WebhooksService struct {
	client *Client
}

// Webhook is a subscription that delivers event notifications to a URL.
type Webhook struct {
	ID     int64  `json:"id"`
	URL    string `json:"url"`
	Active bool   `json:"active"`
	// WebhookType is "email_sending", "campaigns", "audit_log", or "inbound_receiving".
	WebhookType string `json:"webhook_type"`
	// PayloadFormat is "json" or "jsonlines".
	PayloadFormat string `json:"payload_format"`
	// SendingStream is "transactional" or "bulk" (email_sending webhooks only).
	SendingStream  string   `json:"sending_stream,omitempty"`
	DomainID       *int64   `json:"domain_id,omitempty"`
	InboundInboxID *int64   `json:"inbound_inbox_id,omitempty"`
	EventTypes     []string `json:"event_types,omitempty"`
	// SigningSecret verifies payload signatures (HMAC SHA-256). Returned only by
	// Create — store it securely.
	SigningSecret string `json:"signing_secret,omitempty"`
}

// Webhook types for CreateWebhookRequest.WebhookType.
const (
	WebhookTypeEmailSending     = "email_sending"
	WebhookTypeCampaigns        = "campaigns"
	WebhookTypeAuditLog         = "audit_log"
	WebhookTypeInboundReceiving = "inbound_receiving"
)

// Webhook payload formats for CreateWebhookRequest.PayloadFormat.
const (
	PayloadFormatJSON      = "json"
	PayloadFormatJSONLines = "jsonlines"
)

// Webhook event types for CreateWebhookRequest.EventTypes (email_sending and
// campaigns webhooks).
const (
	WebhookEventDelivery      = "delivery"
	WebhookEventSoftBounce    = "soft_bounce"
	WebhookEventBounce        = "bounce"
	WebhookEventSuspension    = "suspension"
	WebhookEventUnsubscribe   = "unsubscribe"
	WebhookEventOpen          = "open"
	WebhookEventSpamComplaint = "spam_complaint"
	WebhookEventClick         = "click"
	WebhookEventReject        = "reject"
)

// CreateWebhookRequest is the payload for creating a webhook. URL and
// WebhookType are required; the rest are optional (Active defaults to true).
type CreateWebhookRequest struct {
	URL            string   `json:"url"`
	WebhookType    string   `json:"webhook_type"`
	Active         *bool    `json:"active,omitempty"`
	PayloadFormat  string   `json:"payload_format,omitempty"`
	SendingStream  string   `json:"sending_stream,omitempty"`
	EventTypes     []string `json:"event_types,omitempty"`
	DomainID       *int64   `json:"domain_id,omitempty"`
	InboundInboxID *int64   `json:"inbound_inbox_id,omitempty"`
}

// UpdateWebhookRequest changes a webhook. Only URL, Active, PayloadFormat, and
// EventTypes are mutable; only the set fields are sent.
type UpdateWebhookRequest struct {
	URL           string   `json:"url,omitempty"`
	Active        *bool    `json:"active,omitempty"`
	PayloadFormat string   `json:"payload_format,omitempty"`
	EventTypes    []string `json:"event_types,omitempty"`
}

// List returns all webhooks for the account.
func (s *WebhooksService) List(ctx context.Context) ([]*Webhook, *Response, error) {
	var wrapper struct {
		Data []*Webhook `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/webhooks", nil, nil, &wrapper)
	return wrapper.Data, resp, err
}

// Get returns a webhook by ID. The signing secret is not included; it is only
// returned by Create.
func (s *WebhooksService) Get(ctx context.Context, webhookID int64) (*Webhook, *Response, error) {
	path := fmt.Sprintf("/api/webhooks/%d", webhookID)
	return s.doWebhook(ctx, http.MethodGet, path, nil)
}

// Create adds a webhook. Its SigningSecret is returned only on creation (never
// by Get or List), so store it to verify payload signatures.
func (s *WebhooksService) Create(ctx context.Context, req *CreateWebhookRequest) (*Webhook, *Response, error) {
	body := map[string]any{"webhook": req}
	return s.doWebhook(ctx, http.MethodPost, "/api/webhooks", body)
}

// Update changes a webhook's mutable fields.
func (s *WebhooksService) Update(ctx context.Context, webhookID int64, req *UpdateWebhookRequest) (*Webhook, *Response, error) {
	path := fmt.Sprintf("/api/webhooks/%d", webhookID)
	body := map[string]any{"webhook": req}
	return s.doWebhook(ctx, http.MethodPatch, path, body)
}

// Delete removes a webhook by ID, returning the deleted webhook.
func (s *WebhooksService) Delete(ctx context.Context, webhookID int64) (*Webhook, *Response, error) {
	path := fmt.Sprintf("/api/webhooks/%d", webhookID)
	return s.doWebhook(ctx, http.MethodDelete, path, nil)
}

// doWebhook sends a request and unwraps the single-webhook data envelope shared
// by Get, Create, Update, and Delete.
func (s *WebhooksService) doWebhook(ctx context.Context, method, path string, body any) (*Webhook, *Response, error) {
	var wrapper struct {
		Data *Webhook `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, method, path, nil, body, &wrapper)
	return wrapper.Data, resp, err
}
