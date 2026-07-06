package mailtrap

import (
	"context"
	"errors"
	"net/http"
	"strconv"
)

// Address is an email address with an optional display name.
type Address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Attachment is a file attached to an outgoing email.
type Attachment struct {
	// Content is the base64-encoded file content.
	Content string `json:"content"`
	// Type is the MIME type, e.g. "text/csv".
	Type string `json:"type,omitempty"`
	// Filename is the attachment's file name.
	Filename string `json:"filename"`
	// Disposition is "attachment" (default) or "inline".
	Disposition string `json:"disposition,omitempty"`
	// ContentID references an inline attachment from the HTML body.
	ContentID string `json:"content_id,omitempty"`
}

// Attachment disposition values for Attachment.Disposition.
const (
	DispositionAttachment = "attachment"
	DispositionInline     = "inline"
)

// SendRequest is an email to send. Provide Text, HTML, or both, or set
// TemplateUUID with TemplateVariables to send from a template.
type SendRequest struct {
	From              Address           `json:"from"`
	To                []Address         `json:"to,omitempty"`
	Cc                []Address         `json:"cc,omitempty"`
	Bcc               []Address         `json:"bcc,omitempty"`
	ReplyTo           *Address          `json:"reply_to,omitempty"`
	Subject           string            `json:"subject,omitempty"`
	Text              string            `json:"text,omitempty"`
	HTML              string            `json:"html,omitempty"`
	Category          string            `json:"category,omitempty"`
	Attachments       []Attachment      `json:"attachments,omitempty"`
	Headers           map[string]string `json:"headers,omitempty"`
	CustomVariables   map[string]any    `json:"custom_variables,omitempty"`
	TemplateUUID      string            `json:"template_uuid,omitempty"`
	TemplateVariables map[string]any    `json:"template_variables,omitempty"`
}

// SendResponse is the result of a successful send.
type SendResponse struct {
	Success    bool     `json:"success"`
	MessageIDs []string `json:"message_ids"`
}

// BatchSendRequest sends many emails in one call. Base holds fields shared by
// every message in Requests; per-request fields override Base.
type BatchSendRequest struct {
	Base     *SendRequest  `json:"base,omitempty"`
	Requests []SendRequest `json:"requests"`
}

// BatchSendResponse is the result of a batch send; Responses is aligned with
// the request order.
type BatchSendResponse struct {
	Success   bool                    `json:"success"`
	Errors    []string                `json:"errors,omitempty"`
	Responses []BatchSendResponseItem `json:"responses"`
}

// BatchSendResponseItem is the per-message result within a batch send.
type BatchSendResponseItem struct {
	Success    bool     `json:"success"`
	MessageIDs []string `json:"message_ids,omitempty"`
	Errors     []string `json:"errors,omitempty"`
}

// Send sends an email. The destination is chosen from the client's
// configuration: the sandbox in sandbox mode (captured, not delivered to real
// recipients), the bulk host in bulk mode, or the transactional host otherwise.
func (c *Client) Send(ctx context.Context, req *SendRequest) (*SendResponse, *Response, error) {
	if req == nil {
		return nil, nil, errors.New("mailtrap: send request must not be nil")
	}
	host, path := c.sendRoute("/api/send")
	out := new(SendResponse)
	resp, err := c.do(ctx, host, http.MethodPost, path, nil, req, out)
	return out, resp, err
}

// SendBatch sends a batch of emails in a single request, routed like Send.
func (c *Client) SendBatch(ctx context.Context, req *BatchSendRequest) (*BatchSendResponse, *Response, error) {
	if req == nil {
		return nil, nil, errors.New("mailtrap: batch send request must not be nil")
	}
	host, path := c.sendRoute("/api/batch")
	out := new(BatchSendResponse)
	resp, err := c.do(ctx, host, http.MethodPost, path, nil, req, out)
	return out, resp, err
}

// sendRoute selects the host and path for sending. Sandbox sends are scoped to
// the configured sandbox ID, e.g. /api/send/123.
func (c *Client) sendRoute(base string) (Host, string) {
	switch {
	case c.sandbox:
		return HostSandbox, base + "/" + strconv.FormatInt(c.sandboxID, 10)
	case c.bulk:
		return HostBulk, base
	default:
		return HostSend, base
	}
}
