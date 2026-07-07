package mailtrap

import (
	"context"
	"fmt"
	"net/http"
)

// EmailTemplatesService manages the account's email templates.
type EmailTemplatesService struct {
	client *Client
}

// EmailTemplate is a reusable email template.
type EmailTemplate struct {
	ID        int64  `json:"id"`
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Subject   string `json:"subject"`
	BodyText  string `json:"body_text,omitempty"`
	BodyHTML  string `json:"body_html,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// EmailTemplateRequest is the payload for creating or updating a template. On
// create, Name, Subject, and Category are required; on update, only the set
// fields are changed.
type EmailTemplateRequest struct {
	Name     string `json:"name,omitempty"`
	Category string `json:"category,omitempty"`
	Subject  string `json:"subject,omitempty"`
	BodyText string `json:"body_text,omitempty"`
	BodyHTML string `json:"body_html,omitempty"`
}

// List returns all email templates for the account.
func (s *EmailTemplatesService) List(ctx context.Context) ([]*EmailTemplate, *Response, error) {
	var templates []*EmailTemplate
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/email_templates", nil, nil, &templates)
	return templates, resp, err
}

// Get returns an email template by ID.
func (s *EmailTemplatesService) Get(ctx context.Context, templateID int64) (*EmailTemplate, *Response, error) {
	path := fmt.Sprintf("/api/email_templates/%d", templateID)
	template := new(EmailTemplate)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, template)
	return template, resp, err
}

// Create adds an email template.
func (s *EmailTemplatesService) Create(ctx context.Context, req *EmailTemplateRequest) (*EmailTemplate, *Response, error) {
	body := map[string]any{"email_template": req}
	template := new(EmailTemplate)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, "/api/email_templates", nil, body, template)
	return template, resp, err
}

// Update changes the set fields of an email template.
func (s *EmailTemplatesService) Update(ctx context.Context, templateID int64, req *EmailTemplateRequest) (*EmailTemplate, *Response, error) {
	path := fmt.Sprintf("/api/email_templates/%d", templateID)
	body := map[string]any{"email_template": req}
	template := new(EmailTemplate)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPatch, path, nil, body, template)
	return template, resp, err
}

// Delete removes an email template by ID.
func (s *EmailTemplatesService) Delete(ctx context.Context, templateID int64) (*Response, error) {
	path := fmt.Sprintf("/api/email_templates/%d", templateID)
	return s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, nil)
}
