package mailtrap

import (
	"context"
	"net/http"
)

// BillingService reads the account's current billing-cycle usage.
type BillingService struct {
	client *Client
}

// BillingUsage reports plan and usage for each product in the current cycle. A
// product is nil when the account has no plan for it.
type BillingUsage struct {
	Billing   *BillingCycle   `json:"billing"`
	Testing   *BillingProduct `json:"testing,omitempty"`
	Sending   *BillingProduct `json:"sending,omitempty"`
	Marketing *BillingProduct `json:"marketing,omitempty"`
}

// BillingCycle is the current billing period.
type BillingCycle struct {
	CycleStart string `json:"cycle_start"`
	CycleEnd   string `json:"cycle_end"`
}

// BillingProduct is one product's plan and usage.
type BillingProduct struct {
	Plan  *BillingPlan         `json:"plan"`
	Usage *BillingUsageMetrics `json:"usage"`
}

// BillingPlan names the active plan for a product.
type BillingPlan struct {
	Name string `json:"name"`
}

// BillingUsageMetrics counts messages used against the plan limit.
// ForwardedMessagesCount is only set for the Sandbox (testing) product.
type BillingUsageMetrics struct {
	SentMessagesCount      *BillingCounter `json:"sent_messages_count,omitempty"`
	ForwardedMessagesCount *BillingCounter `json:"forwarded_messages_count,omitempty"`
}

// BillingCounter is a used/limit pair.
type BillingCounter struct {
	Current int `json:"current"`
	Limit   int `json:"limit"`
}

// Usage returns the account's usage for the current billing cycle across
// Sandbox, Email Sending, and Email Marketing.
func (s *BillingService) Usage(ctx context.Context) (*BillingUsage, *Response, error) {
	usage := new(BillingUsage)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/billing/usage", nil, nil, usage)
	return usage, resp, err
}
