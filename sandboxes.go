package mailtrap

import (
	"context"
	"fmt"
	"net/http"
)

// SandboxesService manages sandboxes (testing inboxes) and their actions.
type SandboxesService struct {
	client *Client
}

// Sandbox is a testing inbox that captures emails sent to it instead of
// delivering them to real recipients.
type Sandbox struct {
	ID                      int64        `json:"id"`
	Name                    string       `json:"name"`
	Username                string       `json:"username"`
	Password                string       `json:"password"`
	MaxSize                 int64        `json:"max_size"`
	Status                  string       `json:"status"`
	EmailUsername           string       `json:"email_username"`
	EmailUsernameEnabled    bool         `json:"email_username_enabled"`
	SentMessagesCount       int64        `json:"sent_messages_count"`
	ForwardedMessagesCount  int64        `json:"forwarded_messages_count"`
	Used                    bool         `json:"used"`
	ForwardFromEmailAddress string       `json:"forward_from_email_address"`
	ProjectID               int64        `json:"project_id"`
	Domain                  string       `json:"domain"`
	POP3Domain              string       `json:"pop3_domain"`
	EmailDomain             string       `json:"email_domain"`
	APIDomain               string       `json:"api_domain"`
	EmailsCount             int64        `json:"emails_count"`
	EmailsUnreadCount       int64        `json:"emails_unread_count"`
	LastMessageSentAt       *string      `json:"last_message_sent_at"`
	SMTPPorts               []int        `json:"smtp_ports"`
	POP3Ports               []int        `json:"pop3_ports"`
	MaxMessageSize          int64        `json:"max_message_size"`
	Permissions             *Permissions `json:"permissions,omitempty"`
}

// SandboxUpdateRequest holds the editable attributes of a sandbox; empty fields
// are omitted from the request.
type SandboxUpdateRequest struct {
	Name          string `json:"name,omitempty"`
	EmailUsername string `json:"email_username,omitempty"`
}

// List returns all sandboxes accessible to the token.
func (s *SandboxesService) List(ctx context.Context) ([]*Sandbox, *Response, error) {
	var sandboxes []*Sandbox
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/sandboxes", nil, nil, &sandboxes)
	return sandboxes, resp, err
}

// Get returns a sandbox by ID.
func (s *SandboxesService) Get(ctx context.Context, sandboxID int64) (*Sandbox, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d", sandboxID)
	sandbox := new(Sandbox)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, sandbox)
	return sandbox, resp, err
}

// Create creates a sandbox with the given name in a project.
func (s *SandboxesService) Create(ctx context.Context, projectID int64, name string) (*Sandbox, *Response, error) {
	path := fmt.Sprintf("/api/projects/%d/sandboxes", projectID)
	body := map[string]any{"sandbox": map[string]string{"name": name}}
	sandbox := new(Sandbox)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, path, nil, body, sandbox)
	return sandbox, resp, err
}

// Update changes a sandbox's name and/or email username.
func (s *SandboxesService) Update(ctx context.Context, sandboxID int64, req *SandboxUpdateRequest) (*Sandbox, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d", sandboxID)
	body := map[string]any{"sandbox": req}
	sandbox := new(Sandbox)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPatch, path, nil, body, sandbox)
	return sandbox, resp, err
}

// Delete removes a sandbox and all its messages, returning the deleted sandbox.
func (s *SandboxesService) Delete(ctx context.Context, sandboxID int64) (*Sandbox, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d", sandboxID)
	sandbox := new(Sandbox)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, sandbox)
	return sandbox, resp, err
}

// Clean deletes all messages from a sandbox.
func (s *SandboxesService) Clean(ctx context.Context, sandboxID int64) (*Sandbox, *Response, error) {
	return s.action(ctx, sandboxID, "clean")
}

// MarkAllRead marks every message in a sandbox as read.
func (s *SandboxesService) MarkAllRead(ctx context.Context, sandboxID int64) (*Sandbox, *Response, error) {
	return s.action(ctx, sandboxID, "all_read")
}

// ResetCredentials regenerates the sandbox's SMTP credentials.
func (s *SandboxesService) ResetCredentials(ctx context.Context, sandboxID int64) (*Sandbox, *Response, error) {
	return s.action(ctx, sandboxID, "reset_credentials")
}

// ToggleEmailAddress enables or disables the sandbox's email address.
func (s *SandboxesService) ToggleEmailAddress(ctx context.Context, sandboxID int64) (*Sandbox, *Response, error) {
	return s.action(ctx, sandboxID, "toggle_email_username")
}

// ResetEmailAddress resets the username of the sandbox's email address.
func (s *SandboxesService) ResetEmailAddress(ctx context.Context, sandboxID int64) (*Sandbox, *Response, error) {
	return s.action(ctx, sandboxID, "reset_email_username")
}

// action runs a PATCH side-effect endpoint that returns the updated sandbox.
func (s *SandboxesService) action(ctx context.Context, sandboxID int64, name string) (*Sandbox, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/%s", sandboxID, name)
	sandbox := new(Sandbox)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPatch, path, nil, nil, sandbox)
	return sandbox, resp, err
}
