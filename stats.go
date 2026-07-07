package mailtrap

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// StatsService reads aggregated email sending statistics.
type StatsService struct {
	client *Client
}

// SendingStats holds delivery, bounce, open, click, and spam counts and rates
// for a period. Rates are fractions in the range 0..1.
type SendingStats struct {
	DeliveryCount int64   `json:"delivery_count"`
	DeliveryRate  float64 `json:"delivery_rate"`
	BounceCount   int64   `json:"bounce_count"`
	BounceRate    float64 `json:"bounce_rate"`
	OpenCount     int64   `json:"open_count"`
	OpenRate      float64 `json:"open_rate"`
	ClickCount    int64   `json:"click_count"`
	ClickRate     float64 `json:"click_rate"`
	SpamCount     int64   `json:"spam_count"`
	SpamRate      float64 `json:"spam_rate"`
}

// DomainStats holds sending stats for a single domain.
type DomainStats struct {
	DomainID int64        `json:"domain_id"`
	Stats    SendingStats `json:"stats"`
}

// CategoryStats holds sending stats for a single category.
type CategoryStats struct {
	Category string       `json:"category"`
	Stats    SendingStats `json:"stats"`
}

// EmailServiceProviderStats holds sending stats for a single ESP.
type EmailServiceProviderStats struct {
	EmailServiceProvider string       `json:"email_service_provider"`
	Stats                SendingStats `json:"stats"`
}

// DateStats holds sending stats for a single date.
type DateStats struct {
	Date  string       `json:"date"`
	Stats SendingStats `json:"stats"`
}

// StatsOptions filters a stats query. StartDate and EndDate (YYYY-MM-DD) are
// required; the remaining fields narrow the results and may be omitted.
type StatsOptions struct {
	StartDate             string
	EndDate               string
	SendingDomainIDs      []int64
	SendingStreams        []string
	Categories            []string
	EmailServiceProviders []string
}

func (o *StatsOptions) values() url.Values {
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
	for _, id := range o.SendingDomainIDs {
		v.Add("domain_ids[]", strconv.FormatInt(id, 10))
	}
	for _, s := range o.SendingStreams {
		v.Add("sending_streams[]", s)
	}
	for _, c := range o.Categories {
		v.Add("categories[]", c)
	}
	for _, esp := range o.EmailServiceProviders {
		v.Add("email_service_providers[]", esp)
	}
	return v
}

// Get returns the account's aggregated sending stats for the period.
func (s *StatsService) Get(ctx context.Context, opts *StatsOptions) (*SendingStats, *Response, error) {
	stats := new(SendingStats)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/stats", opts.values(), nil, stats)
	return stats, resp, err
}

// ByDomain returns sending stats grouped by domain.
func (s *StatsService) ByDomain(ctx context.Context, opts *StatsOptions) ([]*DomainStats, *Response, error) {
	var out []*DomainStats
	resp, err := s.grouped(ctx, "domains", opts, &out)
	return out, resp, err
}

// ByCategory returns sending stats grouped by category.
func (s *StatsService) ByCategory(ctx context.Context, opts *StatsOptions) ([]*CategoryStats, *Response, error) {
	var out []*CategoryStats
	resp, err := s.grouped(ctx, "categories", opts, &out)
	return out, resp, err
}

// ByEmailServiceProvider returns sending stats grouped by ESP.
func (s *StatsService) ByEmailServiceProvider(ctx context.Context, opts *StatsOptions) ([]*EmailServiceProviderStats, *Response, error) {
	var out []*EmailServiceProviderStats
	resp, err := s.grouped(ctx, "email_service_providers", opts, &out)
	return out, resp, err
}

// ByDate returns sending stats grouped by date.
func (s *StatsService) ByDate(ctx context.Context, opts *StatsOptions) ([]*DateStats, *Response, error) {
	var out []*DateStats
	resp, err := s.grouped(ctx, "date", opts, &out)
	return out, resp, err
}

func (s *StatsService) grouped(ctx context.Context, group string, opts *StatsOptions, out any) (*Response, error) {
	return s.client.do(ctx, HostGeneral, http.MethodGet, "/api/stats/"+group, opts.values(), nil, out)
}
