package mailtrap

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

// sandboxMessagesPerPage is the API's fixed page size for message listings.
const sandboxMessagesPerPage = 30

// SandboxMessagesService manages messages captured by a sandbox.
type SandboxMessagesService struct {
	client *Client
}

// SandboxMessage is an email captured by a sandbox.
type SandboxMessage struct {
	ID              int64            `json:"id"`
	SandboxID       int64            `json:"sandbox_id"`
	Subject         string           `json:"subject"`
	SentAt          string           `json:"sent_at"`
	FromEmail       string           `json:"from_email"`
	FromName        string           `json:"from_name"`
	ToEmail         string           `json:"to_email"`
	ToName          string           `json:"to_name"`
	EmailSize       int64            `json:"email_size"`
	IsRead          bool             `json:"is_read"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
	HTMLBodySize    int64            `json:"html_body_size"`
	TextBodySize    int64            `json:"text_body_size"`
	HumanSize       string           `json:"human_size"`
	HTMLPath        string           `json:"html_path"`
	TxtPath         string           `json:"txt_path"`
	RawPath         string           `json:"raw_path"`
	DownloadPath    string           `json:"download_path"`
	HTMLSourcePath  string           `json:"html_source_path"`
	SMTPInformation *SMTPInformation `json:"smtp_information,omitempty"`
}

// SMTPInformation holds the SMTP-level details captured for a message.
type SMTPInformation struct {
	OK   bool                 `json:"ok"`
	Data *SMTPInformationData `json:"data,omitempty"`
}

// SMTPInformationData holds the envelope information of a captured message.
type SMTPInformationData struct {
	MailFromAddr string `json:"mail_from_addr"`
	ClientIP     string `json:"client_ip"`
}

// SpamReport is a message's spam analysis.
type SpamReport struct {
	ResponseCode    int                `json:"ResponseCode"`
	ResponseMessage string             `json:"ResponseMessage"`
	ResponseVersion string             `json:"ResponseVersion"`
	Score           float64            `json:"Score"`
	Spam            bool               `json:"Spam"`
	Threshold       float64            `json:"Threshold"`
	Details         []SpamReportDetail `json:"Details"`
}

// SpamReportDetail is one rule hit in a SpamReport.
type SpamReportDetail struct {
	Pts         float64 `json:"Pts"`
	RuleName    string  `json:"RuleName"`
	Description string  `json:"Description"`
}

// HTMLAnalysisReport is a message's HTML compatibility analysis.
type HTMLAnalysisReport struct {
	Success string              `json:"success"`
	Errors  []HTMLAnalysisError `json:"errors"`
}

// HTMLAnalysisError is one HTML issue found by the analyzer.
type HTMLAnalysisError struct {
	ErrorLine    int                      `json:"error_line"`
	RuleName     string                   `json:"rule_name"`
	EmailClients HTMLAnalysisEmailClients `json:"email_clients"`
}

// HTMLAnalysisEmailClients lists the clients affected by an HTML issue.
type HTMLAnalysisEmailClients struct {
	Desktop []string `json:"desktop"`
	Mobile  []string `json:"mobile"`
	Web     []string `json:"web"`
}

// MessageListOptions are the filters for listing sandbox messages.
type MessageListOptions struct {
	// Search matches the subject, to_email, and to_name.
	Search string
	// Page selects a page of results (max 30 per page).
	Page int
	// LastID returns the page before this message ID, overriding Page.
	LastID int64
}

func (o *MessageListOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	if o.Search != "" {
		v.Set("search", o.Search)
	}
	if o.Page > 0 {
		v.Set("page", strconv.Itoa(o.Page))
	}
	if o.LastID > 0 {
		v.Set("last_id", strconv.FormatInt(o.LastID, 10))
	}
	return v
}

// List returns a page of messages in a sandbox. Response.NextPage is set when
// a full page is returned and more results are likely available.
func (s *SandboxMessagesService) List(ctx context.Context, sandboxID int64, opts *MessageListOptions) ([]*SandboxMessage, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages", sandboxID)
	var messages []*SandboxMessage
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, opts.values(), nil, &messages)
	if err != nil {
		return nil, resp, err
	}
	page := 1
	if opts != nil && opts.Page > 0 {
		page = opts.Page
	}
	if len(messages) >= sandboxMessagesPerPage {
		resp.NextPage = page + 1
	}
	return messages, resp, nil
}

// All iterates every message in a sandbox, fetching pages on demand (Go 1.23
// range-over-func). Iteration stops at the first error, which is yielded once.
func (s *SandboxMessagesService) All(ctx context.Context, sandboxID int64, opts *MessageListOptions) iter.Seq2[*SandboxMessage, error] {
	return func(yield func(*SandboxMessage, error) bool) {
		o := MessageListOptions{}
		if opts != nil {
			o = *opts
		}
		o.LastID = 0 // page-based iteration
		if o.Page <= 0 {
			o.Page = 1
		}
		for {
			messages, _, err := s.List(ctx, sandboxID, &o)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, m := range messages {
				if !yield(m, nil) {
					return
				}
			}
			if len(messages) < sandboxMessagesPerPage {
				return
			}
			o.Page++
		}
	}
}

// Get returns a message by ID.
func (s *SandboxMessagesService) Get(ctx context.Context, sandboxID, messageID int64) (*SandboxMessage, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages/%d", sandboxID, messageID)
	message := new(SandboxMessage)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, message)
	return message, resp, err
}

// Update sets a message's read state, returning the updated message.
func (s *SandboxMessagesService) Update(ctx context.Context, sandboxID, messageID int64, isRead bool) (*SandboxMessage, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages/%d", sandboxID, messageID)
	body := map[string]any{"message": map[string]any{"is_read": isRead}}
	message := new(SandboxMessage)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPatch, path, nil, body, message)
	return message, resp, err
}

// Delete removes a message, returning the deleted message.
func (s *SandboxMessagesService) Delete(ctx context.Context, sandboxID, messageID int64) (*SandboxMessage, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages/%d", sandboxID, messageID)
	message := new(SandboxMessage)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, message)
	return message, resp, err
}

// Forward forwards a message to an email address that has been confirmed by its
// recipient in advance.
func (s *SandboxMessagesService) Forward(ctx context.Context, sandboxID, messageID int64, email string) (*Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages/%d/forward", sandboxID, messageID)
	body := map[string]string{"email": email}
	return s.client.do(ctx, HostGeneral, http.MethodPost, path, nil, body, nil)
}

// SpamReport returns the spam analysis of a message.
func (s *SandboxMessagesService) SpamReport(ctx context.Context, sandboxID, messageID int64) (*SpamReport, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages/%d/spam_report", sandboxID, messageID)
	var wrapper struct {
		Report *SpamReport `json:"report"`
	}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, &wrapper)
	return wrapper.Report, resp, err
}

// HTMLAnalysis returns the HTML compatibility analysis of a message.
func (s *SandboxMessagesService) HTMLAnalysis(ctx context.Context, sandboxID, messageID int64) (*HTMLAnalysisReport, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages/%d/analyze", sandboxID, messageID)
	var wrapper struct {
		Report *HTMLAnalysisReport `json:"report"`
	}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, &wrapper)
	return wrapper.Report, resp, err
}

// Headers returns the mail headers of a message.
func (s *SandboxMessagesService) Headers(ctx context.Context, sandboxID, messageID int64) (map[string]string, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages/%d/mail_headers", sandboxID, messageID)
	var wrapper struct {
		Headers map[string]string `json:"headers"`
	}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, &wrapper)
	return wrapper.Headers, resp, err
}

// Text returns the plain-text body of a message.
func (s *SandboxMessagesService) Text(ctx context.Context, sandboxID, messageID int64) ([]byte, *Response, error) {
	return s.body(ctx, sandboxID, messageID, "body.txt")
}

// Raw returns the raw MIME source of a message.
func (s *SandboxMessagesService) Raw(ctx context.Context, sandboxID, messageID int64) ([]byte, *Response, error) {
	return s.body(ctx, sandboxID, messageID, "body.raw")
}

// HTMLSource returns the HTML source of a message.
func (s *SandboxMessagesService) HTMLSource(ctx context.Context, sandboxID, messageID int64) ([]byte, *Response, error) {
	return s.body(ctx, sandboxID, messageID, "body.htmlsource")
}

// HTML returns the formatted HTML body of a message.
func (s *SandboxMessagesService) HTML(ctx context.Context, sandboxID, messageID int64) ([]byte, *Response, error) {
	return s.body(ctx, sandboxID, messageID, "body.html")
}

// EML returns the message in .eml format.
func (s *SandboxMessagesService) EML(ctx context.Context, sandboxID, messageID int64) ([]byte, *Response, error) {
	return s.body(ctx, sandboxID, messageID, "body.eml")
}

// body fetches one of a message's raw (non-JSON) body representations.
func (s *SandboxMessagesService) body(ctx context.Context, sandboxID, messageID int64, name string) ([]byte, *Response, error) {
	path := fmt.Sprintf("/api/sandboxes/%d/messages/%d/%s", sandboxID, messageID, name)
	return s.client.doRaw(ctx, path)
}
