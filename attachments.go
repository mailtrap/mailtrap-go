package mailtrap

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// SandboxAttachmentsService reads attachments of sandbox messages.
type SandboxAttachmentsService struct {
	client *Client
}

// SandboxAttachment is a file attached to a sandbox message.
type SandboxAttachment struct {
	ID                  int64  `json:"id"`
	MessageID           int64  `json:"message_id"`
	Filename            string `json:"filename"`
	AttachmentType      string `json:"attachment_type"`
	ContentType         string `json:"content_type"`
	ContentID           string `json:"content_id"`
	TransferEncoding    string `json:"transfer_encoding"`
	AttachmentSize      int64  `json:"attachment_size"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	AttachmentHumanSize string `json:"attachment_human_size"`
	DownloadPath        string `json:"download_path"`
}

// AttachmentListOptions filters a sandbox message's attachments.
type AttachmentListOptions struct {
	// Type filters by attachment type, e.g. "inline" or "attachment".
	Type string
}

// List returns the attachments of a message.
func (s *SandboxAttachmentsService) List(ctx context.Context, sandboxID, messageID int64, opts *AttachmentListOptions) ([]*SandboxAttachment, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages/%d/attachments", sandboxID, messageID)
	query := url.Values{}
	if opts != nil && opts.Type != "" {
		query.Set("attachment_type", opts.Type)
	}
	var attachments []*SandboxAttachment
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, query, nil, &attachments)
	return attachments, resp, err
}

// Get returns a single attachment by ID.
func (s *SandboxAttachmentsService) Get(ctx context.Context, sandboxID, messageID, attachmentID int64) (*SandboxAttachment, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages/%d/attachments/%d", sandboxID, messageID, attachmentID)
	attachment := new(SandboxAttachment)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, attachment)
	return attachment, resp, err
}
