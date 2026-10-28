package mailtrap

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// MarketingStatsService reads aggregated Email Marketing statistics across
// campaigns and automations.
type MarketingStatsService struct {
	client *Client
}

// Email Marketing sources for MarketingStatsOptions.Sources.
const (
	MarketingStatsSourceCampaigns   = "campaigns"
	MarketingStatsSourceAutomations = "automations"
)

// MarketingStats holds the Email Marketing counts and rates for a period.
// Rates are fractions in the range 0..1: delivery and bounce rates are
// relative to SentCount, the open, click, spam and unsubscription rates to
// DeliveryCount. Every value is 0 when nothing matched the filters.
type MarketingStats struct {
	DeliveryCount       int64   `json:"delivery_count"`
	DeliveryRate        float64 `json:"delivery_rate"`
	BounceCount         int64   `json:"bounce_count"`
	BounceRate          float64 `json:"bounce_rate"`
	OpenCount           int64   `json:"open_count"`
	OpenRate            float64 `json:"open_rate"`
	ClickCount          int64   `json:"click_count"`
	ClickRate           float64 `json:"click_rate"`
	SpamCount           int64   `json:"spam_count"`
	SpamRate            float64 `json:"spam_rate"`
	UnsubscriptionCount int64   `json:"unsubscription_count"`
	UnsubscriptionRate  float64 `json:"unsubscription_rate"`
	SentCount           int64   `json:"sent_count"`
}

// MarketingStatsOptions filters a marketing stats query. StartDate and EndDate
// (YYYY-MM-DD) are required; the remaining fields narrow the results and may
// be omitted.
type MarketingStatsOptions struct {
	StartDate        string
	EndDate          string
	SendingDomainIDs []int64
	// Sources narrows the stats to MarketingStatsSourceCampaigns or
	// MarketingStatsSourceAutomations; omit it, or pass both, for all Email
	// Marketing traffic.
	Sources               []string
	EmailServiceProviders []string
}

func (o *MarketingStatsOptions) values() url.Values {
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
		v.Add("sending_domain_ids[]", strconv.FormatInt(id, 10))
	}
	for _, s := range o.Sources {
		v.Add("sources[]", s)
	}
	for _, esp := range o.EmailServiceProviders {
		v.Add("email_service_providers[]", esp)
	}
	return v
}

// Get returns the account's aggregated Email Marketing stats for the period:
// the numbers the Email Marketing Stats Overview shows.
func (s *MarketingStatsService) Get(ctx context.Context, opts *MarketingStatsOptions) (*MarketingStats, *Response, error) {
	var wrapper struct {
		Data *MarketingStats `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/marketing/stats", opts.values(), nil, &wrapper)
	return wrapper.Data, resp, err
}
