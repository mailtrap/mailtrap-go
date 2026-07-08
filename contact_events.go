package mailtrap

import (
	"context"
	"net/http"
)

// ContactEventsService records custom events for a contact.
type ContactEventsService struct {
	client *Client
}

// ContactEvent is a custom event recorded against a contact.
type ContactEvent struct {
	ContactID    string `json:"contact_id"`
	ContactEmail string `json:"contact_email"`
	Name         string `json:"name"`
	// Params maps event parameter names to scalar values.
	Params map[string]any `json:"params,omitempty"`
}

// CreateContactEventRequest is the payload for recording an event. Name is
// required (max 255 characters).
type CreateContactEventRequest struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"params,omitempty"`
}

// Create records an event for the contact identified by UUID or email.
func (s *ContactEventsService) Create(ctx context.Context, identifier string, req *CreateContactEventRequest) (*ContactEvent, *Response, error) {
	path := contactPath(identifier) + "/events"
	event := new(ContactEvent)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, path, nil, req, event)
	return event, resp, err
}
