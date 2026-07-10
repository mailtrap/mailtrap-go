package mailtrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// errNoOrganizationID is returned by SubAccounts operations when the client was
// created without WithOrganizationID.
var errNoOrganizationID = errors.New("mailtrap: WithOrganizationID is required for sub-account operations")

// SubAccountsService lists and creates the sub-accounts of an organization.
// These endpoints require an organization token with sub-account management
// permission.
type SubAccountsService struct {
	client *Client
}

// SubAccount is an account within an organization.
type SubAccount struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// List returns the sub-accounts of the organization set with WithOrganizationID.
func (s *SubAccountsService) List(ctx context.Context) ([]*SubAccount, *Response, error) {
	if s.client.organizationID == 0 {
		return nil, nil, errNoOrganizationID
	}
	path := fmt.Sprintf("/api/organizations/%d/sub_accounts", s.client.organizationID)
	var subAccounts []*SubAccount
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, &subAccounts)
	return subAccounts, resp, err
}

// Create adds a sub-account with the given name under the organization set with
// WithOrganizationID.
func (s *SubAccountsService) Create(ctx context.Context, name string) (*SubAccount, *Response, error) {
	if s.client.organizationID == 0 {
		return nil, nil, errNoOrganizationID
	}
	path := fmt.Sprintf("/api/organizations/%d/sub_accounts", s.client.organizationID)
	body := map[string]any{"account": map[string]string{"name": name}}
	subAccount := new(SubAccount)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, path, nil, body, subAccount)
	return subAccount, resp, err
}
