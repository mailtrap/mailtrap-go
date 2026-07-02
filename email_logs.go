package mailtrap

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
)

// EmailLogsService reads the account's email logs (sent messages and their
// delivery events).
type EmailLogsService struct {
	client *Client
}

// EmailLogMessage is a logged message. List populates the summary fields; Get
// also fills RawMessageURL and Events.
type EmailLogMessage struct {
	MessageID string `json:"message_id"`
	// Status is "delivered", "not_delivered", "enqueued", or "opted_out".
	Status            string         `json:"status"`
	Subject           string         `json:"subject"`
	From              string         `json:"from"`
	To                string         `json:"to"`
	SentAt            string         `json:"sent_at"`
	ClientIP          string         `json:"client_ip"`
	Category          string         `json:"category"`
	CustomVariables   map[string]any `json:"custom_variables"`
	SendingStream     string         `json:"sending_stream"`
	DomainID          int64          `json:"domain_id"`
	TemplateID        *int64         `json:"template_id"`
	TemplateVariables map[string]any `json:"template_variables"`
	OpensCount        int64          `json:"opens_count"`
	ClicksCount       int64          `json:"clicks_count"`
	// RawMessageURL is a temporary signed URL to the raw .eml (Get only).
	RawMessageURL string `json:"raw_message_url,omitempty"`
	// Events is the delivery-lifecycle history (Get only).
	Events []MessageEvent `json:"events,omitempty"`
}

// Email log message statuses for EmailLogMessage.Status.
const (
	EmailLogStatusDelivered    = "delivered"
	EmailLogStatusNotDelivered = "not_delivered"
	EmailLogStatusEnqueued     = "enqueued"
	EmailLogStatusOptedOut     = "opted_out"
)

// MessageEvent is one delivery-lifecycle event. Which Details fields are
// populated depends on EventType (delivery, open, click, soft_bounce, bounce,
// spam, unsubscribe, suspension, reject).
type MessageEvent struct {
	EventType string              `json:"event_type"`
	CreatedAt string              `json:"created_at"`
	Details   MessageEventDetails `json:"details"`
}

// Message event types for MessageEvent.EventType.
const (
	MessageEventTypeDelivery    = "delivery"
	MessageEventTypeOpen        = "open"
	MessageEventTypeClick       = "click"
	MessageEventTypeSoftBounce  = "soft_bounce"
	MessageEventTypeBounce      = "bounce"
	MessageEventTypeSpam        = "spam"
	MessageEventTypeUnsubscribe = "unsubscribe"
	MessageEventTypeSuspension  = "suspension"
	MessageEventTypeReject      = "reject"
)

// MessageEventDetails is the union of every event's detail fields; only those
// relevant to the event type are set.
type MessageEventDetails struct {
	SendingIP                    string `json:"sending_ip,omitempty"`
	RecipientMX                  string `json:"recipient_mx,omitempty"`
	EmailServiceProvider         string `json:"email_service_provider,omitempty"`
	EmailServiceProviderStatus   string `json:"email_service_provider_status,omitempty"`
	EmailServiceProviderResponse string `json:"email_service_provider_response,omitempty"`
	BounceCategory               string `json:"bounce_category,omitempty"`
	WebIPAddress                 string `json:"web_ip_address,omitempty"`
	ClickURL                     string `json:"click_url,omitempty"`
	SpamFeedbackType             string `json:"spam_feedback_type,omitempty"`
	RejectReason                 string `json:"reject_reason,omitempty"`
}

// EmailLogsList is a page of email logs. NextPageCursor is empty on the last page.
type EmailLogsList struct {
	Messages       []*EmailLogMessage `json:"messages"`
	TotalCount     int64              `json:"total_count"`
	NextPageCursor string             `json:"next_page_cursor"`
}

// LogFilter is one email-logs filter: a comparison operator and its value(s).
// A single value is sent as a scalar and multiple values as an array; the
// empty and not_empty operators take no value.
type LogFilter struct {
	Operator string
	Values   []string
}

// EmailLogsListOptions filters and paginates an email-logs listing. Filters
// maps a field name (e.g. "to", "status", "category", "sending_domain_id") to
// its filter; SentAfter/SentBefore bound sent_at (ISO 8601).
type EmailLogsListOptions struct {
	// SearchAfter is the NextPageCursor from a previous response.
	SearchAfter string
	SentAfter   string
	SentBefore  string
	Filters     map[string]LogFilter
}

func (o *EmailLogsListOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	if o.SearchAfter != "" {
		v.Set("search_after", o.SearchAfter)
	}
	if o.SentAfter != "" {
		v.Set("filters[sent_after]", o.SentAfter)
	}
	if o.SentBefore != "" {
		v.Set("filters[sent_before]", o.SentBefore)
	}
	for field, f := range o.Filters {
		v.Set(fmt.Sprintf("filters[%s][operator]", field), f.Operator)
		switch len(f.Values) {
		case 0:
		case 1:
			v.Set(fmt.Sprintf("filters[%s][value]", field), f.Values[0])
		default:
			key := fmt.Sprintf("filters[%s][value][]", field)
			for _, val := range f.Values {
				v.Add(key, val)
			}
		}
	}
	return v
}

// List returns a page of email logs matching opts, ordered by sent_at
// descending. Follow EmailLogsList.NextPageCursor with SearchAfter for the
// next page, or use All to iterate every match.
func (s *EmailLogsService) List(ctx context.Context, opts *EmailLogsListOptions) (*EmailLogsList, *Response, error) {
	list := new(EmailLogsList)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/email_logs", opts.values(), nil, list)
	return list, resp, err
}

// All iterates every email log matching opts, following the cursor across pages.
// Iteration stops at the first error, yielded once.
func (s *EmailLogsService) All(ctx context.Context, opts *EmailLogsListOptions) iter.Seq2[*EmailLogMessage, error] {
	return func(yield func(*EmailLogMessage, error) bool) {
		o := EmailLogsListOptions{}
		if opts != nil {
			o = *opts
		}
		for {
			list, _, err := s.List(ctx, &o)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, m := range list.Messages {
				if !yield(m, nil) {
					return
				}
			}
			if list.NextPageCursor == "" {
				return
			}
			o.SearchAfter = list.NextPageCursor
		}
	}
}

// Get returns a single email log message by its UUID, including its events.
func (s *EmailLogsService) Get(ctx context.Context, messageID string) (*EmailLogMessage, *Response, error) {
	path := fmt.Sprintf("/api/email_logs/%s", messageID)
	msg := new(EmailLogMessage)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, msg)
	return msg, resp, err
}
