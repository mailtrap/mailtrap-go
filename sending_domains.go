package mailtrap

import (
	"context"
	"fmt"
	"net/http"
)

// SendingDomainsService manages sending domains and their compliance settings.
type SendingDomainsService struct {
	client *Client
}

// DNSRecord is one DNS record a domain must publish for authentication.
type DNSRecord struct {
	Key    string `json:"key"`
	Domain string `json:"domain"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type"`
	Value  string `json:"value"`
}

// SendingDomainPermissions describes the caller's permissions on a domain.
type SendingDomainPermissions struct {
	CanRead    bool `json:"can_read"`
	CanUpdate  bool `json:"can_update"`
	CanDestroy bool `json:"can_destroy"`
}

// SendingDomain is a domain used for email authentication and sending.
type SendingDomain struct {
	ID                          int64                     `json:"id"`
	DomainName                  string                    `json:"domain_name"`
	Demo                        bool                      `json:"demo"`
	ComplianceStatus            string                    `json:"compliance_status"`
	DNSVerified                 bool                      `json:"dns_verified"`
	DNSVerifiedAt               string                    `json:"dns_verified_at"`
	DNSRecords                  []DNSRecord               `json:"dns_records"`
	OpenTrackingEnabled         bool                      `json:"open_tracking_enabled"`
	ClickTrackingEnabled        bool                      `json:"click_tracking_enabled"`
	AutoUnsubscribeLinkEnabled  bool                      `json:"auto_unsubscribe_link_enabled"`
	CustomDomainTrackingEnabled bool                      `json:"custom_domain_tracking_enabled"`
	HealthAlertsEnabled         bool                      `json:"health_alerts_enabled"`
	CriticalAlertsEnabled       bool                      `json:"critical_alerts_enabled"`
	AlertRecipientEmail         string                    `json:"alert_recipient_email"`
	Permissions                 *SendingDomainPermissions `json:"permissions,omitempty"`
}

// UpdateDomainRequest changes a domain's tracking settings. Only the non-nil
// fields are sent, so leave a field nil to keep its current value.
type UpdateDomainRequest struct {
	OpenTrackingEnabled        *bool `json:"open_tracking_enabled,omitempty"`
	ClickTrackingEnabled       *bool `json:"click_tracking_enabled,omitempty"`
	AutoUnsubscribeLinkEnabled *bool `json:"auto_unsubscribe_link_enabled,omitempty"`
}

// CompanyInfo is the sender's company information used for compliance review.
type CompanyInfo struct {
	Name              string `json:"name"`
	Address           string `json:"address"`
	City              string `json:"city"`
	Country           string `json:"country"`
	Phone             string `json:"phone"`
	ZipCode           string `json:"zip_code"`
	PrivacyPolicyURL  string `json:"privacy_policy_url"`
	TermsOfServiceURL string `json:"terms_of_service_url"`
	WebsiteURL        string `json:"website_url"`
	// InfoLevel is "business" or "individual".
	InfoLevel string `json:"info_level"`
}

// CompanyInfoRequest is the payload for creating or updating company info. On
// create, name, address, city, country, zip_code, and website_url are required;
// on update, every field is optional and only the set fields are changed.
type CompanyInfoRequest struct {
	Name              string `json:"name,omitempty"`
	Address           string `json:"address,omitempty"`
	City              string `json:"city,omitempty"`
	Country           string `json:"country,omitempty"`
	ZipCode           string `json:"zip_code,omitempty"`
	WebsiteURL        string `json:"website_url,omitempty"`
	Phone             string `json:"phone,omitempty"`
	PrivacyPolicyURL  string `json:"privacy_policy_url,omitempty"`
	TermsOfServiceURL string `json:"terms_of_service_url,omitempty"`
	// InfoLevel is "business" or "individual".
	InfoLevel string `json:"info_level,omitempty"`
}

// Company info levels for CompanyInfo.InfoLevel and CompanyInfoRequest.InfoLevel.
const (
	InfoLevelBusiness   = "business"
	InfoLevelIndividual = "individual"
)

// List returns all sending domains and their verification status.
func (s *SendingDomainsService) List(ctx context.Context) ([]*SendingDomain, *Response, error) {
	var wrapper struct {
		Data []*SendingDomain `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/domains", nil, nil, &wrapper)
	return wrapper.Data, resp, err
}

// Get returns a sending domain by ID.
func (s *SendingDomainsService) Get(ctx context.Context, domainID int64) (*SendingDomain, *Response, error) {
	path := fmt.Sprintf("/api/domains/%d", domainID)
	domain := new(SendingDomain)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, domain)
	return domain, resp, err
}

// Create adds a sending domain. After creation, publish the returned DNS records
// and verify them before sending in production.
func (s *SendingDomainsService) Create(ctx context.Context, domainName string) (*SendingDomain, *Response, error) {
	body := map[string]any{"domain": map[string]string{"domain_name": domainName}}
	domain := new(SendingDomain)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, "/api/domains", nil, body, domain)
	return domain, resp, err
}

// Update changes a domain's tracking settings.
func (s *SendingDomainsService) Update(ctx context.Context, domainID int64, req *UpdateDomainRequest) (*SendingDomain, *Response, error) {
	path := fmt.Sprintf("/api/domains/%d", domainID)
	body := map[string]any{"domain": req}
	domain := new(SendingDomain)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPatch, path, nil, body, domain)
	return domain, resp, err
}

// Delete removes a sending domain by ID.
func (s *SendingDomainsService) Delete(ctx context.Context, domainID int64) (*Response, error) {
	path := fmt.Sprintf("/api/domains/%d", domainID)
	return s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, nil)
}

// SendSetupInstructions emails the domain's DNS setup instructions to an address.
func (s *SendingDomainsService) SendSetupInstructions(ctx context.Context, domainID int64, email string) (*Response, error) {
	path := fmt.Sprintf("/api/domains/%d/send_setup_instructions", domainID)
	body := map[string]string{"email": email}
	return s.client.do(ctx, HostGeneral, http.MethodPost, path, nil, body, nil)
}

// CompanyInfo returns the company information attached to a domain.
func (s *SendingDomainsService) CompanyInfo(ctx context.Context, domainID int64) (*CompanyInfo, *Response, error) {
	return s.companyInfo(ctx, http.MethodGet, domainID, nil)
}

// CreateCompanyInfo sets the company information for a domain.
func (s *SendingDomainsService) CreateCompanyInfo(ctx context.Context, domainID int64, req *CompanyInfoRequest) (*CompanyInfo, *Response, error) {
	return s.companyInfo(ctx, http.MethodPost, domainID, req)
}

// UpdateCompanyInfo changes the set fields of a domain's company information.
func (s *SendingDomainsService) UpdateCompanyInfo(ctx context.Context, domainID int64, req *CompanyInfoRequest) (*CompanyInfo, *Response, error) {
	return s.companyInfo(ctx, http.MethodPatch, domainID, req)
}

func (s *SendingDomainsService) companyInfo(ctx context.Context, method string, domainID int64, req *CompanyInfoRequest) (*CompanyInfo, *Response, error) {
	path := fmt.Sprintf("/api/domains/%d/company_info", domainID)
	var body any
	if req != nil {
		body = map[string]any{"company_info": req}
	}
	var wrapper struct {
		Data *CompanyInfo `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, method, path, nil, body, &wrapper)
	return wrapper.Data, resp, err
}
