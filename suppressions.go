package mailtrap

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// SuppressionsService manages the account's suppression list: addresses that
// Mailtrap will not send to because of bounces, complaints, or unsubscribes.
type SuppressionsService struct {
	client *Client
}

// Suppression is a single suppressed recipient and the message that caused it.
type Suppression struct {
	ID string `json:"id"`
	// Type is the reason: "hard bounce", "unsubscription", "spam complaint",
	// or "manual import".
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	Email     string `json:"email"`
	// SendingStream is "transactional", "bulk", or "any".
	SendingStream          string `json:"sending_stream"`
	DomainName             string `json:"domain_name"`
	MessageBounceCategory  string `json:"message_bounce_category"`
	MessageCategory        string `json:"message_category"`
	MessageClientIP        string `json:"message_client_ip"`
	MessageCreatedAt       string `json:"message_created_at"`
	MessageESPResponse     string `json:"message_esp_response"`
	MessageESPServerType   string `json:"message_esp_server_type"`
	MessageOutgoingIP      string `json:"message_outgoing_ip"`
	MessageRecipientMXName string `json:"message_recipient_mx_name"`
	MessageSenderEmail     string `json:"message_sender_email"`
	MessageSubject         string `json:"message_subject"`
}

// SuppressionListOptions filters a suppression listing. The endpoint returns up
// to 1000 suppressions per request; page through larger lists by passing the
// last returned suppression's ID as LastID.
type SuppressionListOptions struct {
	// Email filters by exact address (case-insensitive).
	Email string
	// StartTime and EndTime bound the created_at timestamp (ISO 8601).
	StartTime string
	EndTime   string
	// LastID returns suppressions after this one, for cursor-based pagination.
	LastID string
}

func (o *SuppressionListOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	if o.Email != "" {
		v.Set("email", o.Email)
	}
	if o.StartTime != "" {
		v.Set("start_time", o.StartTime)
	}
	if o.EndTime != "" {
		v.Set("end_time", o.EndTime)
	}
	if o.LastID != "" {
		v.Set("last_id", o.LastID)
	}
	return v
}

// CreateSuppressionRequest is the payload for suppressing an address. Type is
// optional and defaults to "manual import".
type CreateSuppressionRequest struct {
	Email         string `json:"email"`
	DomainID      int64  `json:"domain_id"`
	SendingStream string `json:"sending_stream"`
	Type          string `json:"type,omitempty"`
}

// List returns suppressions matching opts (pass nil for no filters).
func (s *SuppressionsService) List(ctx context.Context, opts *SuppressionListOptions) ([]*Suppression, *Response, error) {
	var suppressions []*Suppression
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/suppressions", opts.values(), nil, &suppressions)
	return suppressions, resp, err
}

// Create suppresses an address, returning the created suppression.
func (s *SuppressionsService) Create(ctx context.Context, req *CreateSuppressionRequest) (*Suppression, *Response, error) {
	var wrapper struct {
		Data *Suppression `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, "/api/suppressions", nil, req, &wrapper)
	return wrapper.Data, resp, err
}

// Delete removes a suppression by ID, returning the deleted suppression.
// Mailtrap will send to the address again unless it is suppressed anew.
func (s *SuppressionsService) Delete(ctx context.Context, suppressionID string) (*Suppression, *Response, error) {
	path := fmt.Sprintf("/api/suppressions/%s", suppressionID)
	suppression := new(Suppression)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, suppression)
	return suppression, resp, err
}
