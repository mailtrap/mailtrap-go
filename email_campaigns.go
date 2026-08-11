package mailtrap

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

// EmailCampaignsService manages email marketing campaigns: drafting, the
// schedule/start lifecycle, and performance statistics.
type EmailCampaignsService struct {
	client *Client
}

// EmailCampaign is an email marketing campaign. Template.BodyHTML and
// Template.BodyText are returned only on single-campaign responses; the list
// endpoint omits them.
type EmailCampaign struct {
	ID int64 `json:"id"`
	// DomainID is the ID of the sending domain used for the campaign, as
	// returned by the Sending Domains endpoints.
	DomainID        int64                 `json:"domain_id"`
	DomainName      string                `json:"domain_name"`
	Name            string                `json:"name"`
	FromLocalPart   string                `json:"from_local_part"`
	FromDisplayName string                `json:"from_display_name"`
	ReplyTo         *EmailCampaignReplyTo `json:"reply_to"`
	// CurrentState is the campaign's position in its lifecycle.
	CurrentState         string                     `json:"current_state"`
	CurrentStateMetadata EmailCampaignStateMetadata `json:"current_state_metadata"`
	CreatedAt            string                     `json:"created_at"`
	UpdatedAt            string                     `json:"updated_at"`
	LastStartedAt        *string                    `json:"last_started_at"`
	// LastStartedAtDate is the date the campaign was last started; present only
	// when the campaign has been started.
	LastStartedAtDate string `json:"last_started_at_date,omitempty"`
	// RecipientTotalCount is the total number of recipients targeted by the
	// campaign, or nil until the audience is resolved.
	RecipientTotalCount *int64  `json:"recipient_total_count"`
	ContactListIDs      []int64 `json:"contact_list_ids"`
	ContactSegmentIDs   []int64 `json:"contact_segment_ids"`
	// DeliveryMode is how the campaign is delivered: rapid sends as fast as
	// possible; gradual throttles to DeliveryOptions.EmailsPerHour.
	DeliveryMode    string                        `json:"delivery_mode"`
	DeliveryOptions *EmailCampaignDeliveryOptions `json:"delivery_options"`
	Template        *EmailCampaignTemplate        `json:"template"`
}

// Email campaign states for EmailCampaign.CurrentState.
const (
	EmailCampaignStateDraft             = "draft"
	EmailCampaignStateScheduled         = "scheduled"
	EmailCampaignStateStarted           = "started"
	EmailCampaignStateQueued            = "queued"
	EmailCampaignStatePaused            = "paused"
	EmailCampaignStateTerminating       = "terminating"
	EmailCampaignStateUnderReview       = "under_review"
	EmailCampaignStateFinished          = "finished"
	EmailCampaignStateFailed            = "failed"
	EmailCampaignStateFailedImmediately = "failed_immediately"
)

// Email campaign delivery modes for EmailCampaign.DeliveryMode.
const (
	EmailCampaignDeliveryModeRapid   = "rapid"
	EmailCampaignDeliveryModeGradual = "gradual"
)

// EmailCampaignReplyTo holds the parts of a campaign's Reply-To address.
type EmailCampaignReplyTo struct {
	DisplayName string `json:"display_name,omitempty"`
	LocalPart   string `json:"local_part,omitempty"`
	Domain      string `json:"domain,omitempty"`
}

// EmailCampaignStateMetadata describes the campaign's most recent state
// transition. Which fields are set depends on the state.
type EmailCampaignStateMetadata struct {
	Reason string `json:"reason,omitempty"`
	// Error is the last error message recorded for a failed campaign.
	Error string `json:"error,omitempty"`
	// Errors lists per-recipient errors recorded when sending failed.
	Errors []EmailCampaignStateError `json:"errors,omitempty"`
	// ScheduledAt is when the campaign is scheduled to send (ISO 8601); present
	// in the scheduled state.
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

// EmailCampaignStateError is one per-recipient error recorded when sending
// failed.
type EmailCampaignStateError struct {
	Message   string `json:"message"`
	RcptIndex int    `json:"rcpt_index"`
}

// EmailCampaignDeliveryOptions holds delivery throttling options; they apply
// when the campaign's delivery mode is gradual.
type EmailCampaignDeliveryOptions struct {
	EmailsPerHour *int `json:"emails_per_hour,omitempty"`
}

// EmailCampaignTemplateAttributes is the campaign's inline template (subject
// and design) in create and update requests. Updates are partial: only the
// set sub-fields change.
type EmailCampaignTemplateAttributes struct {
	// Subject is the email subject line; required when creating a campaign.
	// Supports merge tags, e.g. "Hi {{first_name}}".
	Subject string `json:"subject,omitempty"`
	// BodyHTML is the HTML body (the design). Optional for a draft; required
	// before the campaign can be scheduled or started. Include an unsubscribe
	// link via an anchor whose href contains the __unsubscribe_url__ placeholder.
	BodyHTML string `json:"body_html,omitempty"`
	// BodyText is an optional plain-text alternative of the email body.
	BodyText string `json:"body_text,omitempty"`
	// MergeTags is the bare names of the merge tags referenced in the
	// subject/body, without the {{ }} delimiters.
	MergeTags []string `json:"merge_tags,omitempty"`
}

// EmailCampaignTemplate is the campaign's template as returned by the API.
type EmailCampaignTemplate struct {
	ID        int64    `json:"id"`
	Subject   string   `json:"subject"`
	MergeTags []string `json:"merge_tags"`
	// BodyHTML is returned only on single-campaign responses.
	BodyHTML *string `json:"body_html,omitempty"`
	// BodyText is returned only on single-campaign responses.
	BodyText *string `json:"body_text,omitempty"`
}

// CreateEmailCampaignRequest is the payload for creating a campaign. Name,
// DomainID, FromLocalPart, and TemplateAttributes.Subject are required; the
// rest are optional. The campaign is always created in the draft state —
// scheduling and starting are separate actions.
type CreateEmailCampaignRequest struct {
	Name string `json:"name"`
	// DomainID is the ID of the verified sending domain to use, as returned by
	// the Sending Domains endpoints.
	DomainID           int64                            `json:"domain_id"`
	FromDisplayName    string                           `json:"from_display_name,omitempty"`
	FromLocalPart      string                           `json:"from_local_part"`
	ReplyTo            *EmailCampaignReplyTo            `json:"reply_to,omitempty"`
	TemplateAttributes *EmailCampaignTemplateAttributes `json:"template_attributes"`
	// DeliveryMode is rapid or gradual.
	DeliveryMode    string                        `json:"delivery_mode,omitempty"`
	DeliveryOptions *EmailCampaignDeliveryOptions `json:"delivery_options,omitempty"`
	// ContactListIDs is the full set of contact lists to send to; lists not
	// listed are removed.
	ContactListIDs []int64 `json:"contact_list_ids,omitempty"`
	// ContactSegmentIDs is the full set of contact segments to send to.
	ContactSegmentIDs []int64 `json:"contact_segment_ids,omitempty"`
}

// UpdateEmailCampaignRequest changes a draft campaign. All fields are
// optional; only the set fields change.
type UpdateEmailCampaignRequest struct {
	Name string `json:"name,omitempty"`
	// DomainID is the ID of the verified sending domain to use, as returned by
	// the Sending Domains endpoints.
	DomainID           int64                            `json:"domain_id,omitempty"`
	FromDisplayName    string                           `json:"from_display_name,omitempty"`
	FromLocalPart      string                           `json:"from_local_part,omitempty"`
	ReplyTo            *EmailCampaignReplyTo            `json:"reply_to,omitempty"`
	TemplateAttributes *EmailCampaignTemplateAttributes `json:"template_attributes,omitempty"`
	// DeliveryMode is rapid or gradual.
	DeliveryMode    string                        `json:"delivery_mode,omitempty"`
	DeliveryOptions *EmailCampaignDeliveryOptions `json:"delivery_options,omitempty"`
	// ContactListIDs is the full set of contact lists to send to; lists not
	// listed are removed. Nil leaves the lists unchanged; a pointer to an
	// empty slice removes them all.
	ContactListIDs *[]int64 `json:"contact_list_ids,omitempty"`
	// ContactSegmentIDs is the full set of contact segments to send to. Nil
	// leaves the segments unchanged; a pointer to an empty slice removes them
	// all.
	ContactSegmentIDs *[]int64 `json:"contact_segment_ids,omitempty"`
}

// EmailCampaignListOptions paginates and filters a campaign listing.
type EmailCampaignListOptions struct {
	// Token is the page number to retrieve (page-token pagination); zero means
	// the first page.
	Token int
	// PerPage is the number of campaigns per page (default 50, maximum 100).
	PerPage int
	// Search filters campaigns by name.
	Search string
}

func (o *EmailCampaignListOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	if o.Token > 0 {
		v.Set("token", strconv.Itoa(o.Token))
	}
	if o.PerPage > 0 {
		v.Set("per_page", strconv.Itoa(o.PerPage))
	}
	if o.Search != "" {
		v.Set("search", o.Search)
	}
	return v
}

// EmailCampaignsList is a page of email campaigns with pagination metadata.
type EmailCampaignsList struct {
	Data       []*EmailCampaign `json:"data"`
	Pagination Pagination       `json:"pagination"`
}

// List returns a page of email campaigns matching opts (pass nil for the
// first page with defaults), newest first. Follow Pagination.NextToken with
// Token for the next page, or use All to iterate every match.
func (s *EmailCampaignsService) List(ctx context.Context, opts *EmailCampaignListOptions) (*EmailCampaignsList, *Response, error) {
	list := new(EmailCampaignsList)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/email_campaigns", opts.values(), nil, list)
	return list, resp, err
}

// All iterates every email campaign matching opts, following the page token
// across pages. Iteration stops at the first error, yielded once.
func (s *EmailCampaignsService) All(ctx context.Context, opts *EmailCampaignListOptions) iter.Seq2[*EmailCampaign, error] {
	return func(yield func(*EmailCampaign, error) bool) {
		o := EmailCampaignListOptions{}
		if opts != nil {
			o = *opts
		}
		for {
			list, _, err := s.List(ctx, &o)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, c := range list.Data {
				if !yield(c, nil) {
					return
				}
			}
			if list.Pagination.NextToken == nil {
				return
			}
			o.Token = *list.Pagination.NextToken
		}
	}
}

// Get returns an email campaign by ID.
func (s *EmailCampaignsService) Get(ctx context.Context, campaignID int64) (*EmailCampaign, *Response, error) {
	path := fmt.Sprintf("/api/email_campaigns/%d", campaignID)
	return s.doCampaign(ctx, http.MethodGet, path, nil)
}

// Create creates an email campaign in the draft state.
func (s *EmailCampaignsService) Create(ctx context.Context, req *CreateEmailCampaignRequest) (*EmailCampaign, *Response, error) {
	return s.doCampaign(ctx, http.MethodPost, "/api/email_campaigns", req)
}

// Update changes a draft campaign; only the set fields change. Updating a
// campaign that is no longer a draft fails with a ValidationError.
func (s *EmailCampaignsService) Update(ctx context.Context, campaignID int64, req *UpdateEmailCampaignRequest) (*EmailCampaign, *Response, error) {
	path := fmt.Sprintf("/api/email_campaigns/%d", campaignID)
	return s.doCampaign(ctx, http.MethodPatch, path, req)
}

// Delete removes an email campaign by ID. The campaign must not be in a
// sending state.
func (s *EmailCampaignsService) Delete(ctx context.Context, campaignID int64) (*Response, error) {
	path := fmt.Sprintf("/api/email_campaigns/%d", campaignID)
	return s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, nil)
}

// Start starts sending a draft campaign immediately.
func (s *EmailCampaignsService) Start(ctx context.Context, campaignID int64) (*EmailCampaign, *Response, error) {
	return s.action(ctx, campaignID, "start", nil)
}

// Schedule schedules a draft campaign to start sending at datetime (ISO 8601,
// in the future and no more than one month ahead). The time is reported back
// in CurrentStateMetadata.ScheduledAt.
func (s *EmailCampaignsService) Schedule(ctx context.Context, campaignID int64, datetime string) (*EmailCampaign, *Response, error) {
	return s.action(ctx, campaignID, "schedule", map[string]string{"datetime": datetime})
}

// Cancel cancels a scheduled campaign, returning it to the draft state.
func (s *EmailCampaignsService) Cancel(ctx context.Context, campaignID int64) (*EmailCampaign, *Response, error) {
	return s.action(ctx, campaignID, "cancel", nil)
}

// Terminate aborts a campaign that is currently sending (started, queued, or
// paused).
func (s *EmailCampaignsService) Terminate(ctx context.Context, campaignID int64) (*EmailCampaign, *Response, error) {
	return s.action(ctx, campaignID, "terminate", nil)
}

// Reset returns a scheduled campaign to the draft state.
func (s *EmailCampaignsService) Reset(ctx context.Context, campaignID int64) (*EmailCampaign, *Response, error) {
	return s.action(ctx, campaignID, "reset", nil)
}

// action runs a POST lifecycle endpoint that returns the updated campaign.
func (s *EmailCampaignsService) action(ctx context.Context, campaignID int64, name string, body any) (*EmailCampaign, *Response, error) {
	path := fmt.Sprintf("/api/email_campaigns/%d/%s", campaignID, name)
	return s.doCampaign(ctx, http.MethodPost, path, body)
}

// doCampaign sends a request and unwraps the single-campaign data envelope
// shared by Get, Create, Update, and the lifecycle actions.
func (s *EmailCampaignsService) doCampaign(ctx context.Context, method, path string, body any) (*EmailCampaign, *Response, error) {
	var wrapper struct {
		Data *EmailCampaign `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, method, path, nil, body, &wrapper)
	return wrapper.Data, resp, err
}

// EmailCampaignStatsOptions narrows the stats aggregation window. Dates use
// the YYYY-MM-DD format; both are optional — the window defaults to the
// period since the campaign was last started.
type EmailCampaignStatsOptions struct {
	// StartDate is the start of the window (inclusive).
	StartDate string
	// EndDate is the end of the window (inclusive); defaults to today.
	EndDate string
}

func (o *EmailCampaignStatsOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	if o.StartDate != "" {
		v.Set("start_date", o.StartDate)
	}
	if o.EndDate != "" {
		v.Set("end_date", o.EndDate)
	}
	return v
}

// EmailCampaignStats holds aggregated campaign performance metrics. Counts
// and rates are 0 when the campaign has not been started; rates are fractions
// in the range 0..1.
type EmailCampaignStats struct {
	DeliveryCount       int64   `json:"delivery_count"`
	OpenCount           int64   `json:"open_count"`
	ClickCount          int64   `json:"click_count"`
	BounceCount         int64   `json:"bounce_count"`
	UnsubscriptionCount int64   `json:"unsubscription_count"`
	SentCount           int64   `json:"sent_count"`
	SpamCount           int64   `json:"spam_count"`
	DeliveryRate        float64 `json:"delivery_rate"`
	OpenRate            float64 `json:"open_rate"`
	ClickRate           float64 `json:"click_rate"`
	BounceRate          float64 `json:"bounce_rate"`
	SpamRate            float64 `json:"spam_rate"`
	UnsubscriptionRate  float64 `json:"unsubscription_rate"`
}

// Stats returns the campaign's aggregated performance metrics, optionally
// narrowed to a date window (pass nil for the default window).
func (s *EmailCampaignsService) Stats(ctx context.Context, campaignID int64, opts *EmailCampaignStatsOptions) (*EmailCampaignStats, *Response, error) {
	path := fmt.Sprintf("/api/email_campaigns/%d/stats", campaignID)
	var wrapper struct {
		Data *EmailCampaignStats `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, opts.values(), nil, &wrapper)
	return wrapper.Data, resp, err
}
