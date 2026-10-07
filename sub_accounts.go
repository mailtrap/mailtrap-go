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

// Delete permanently removes a sub-account of the organization set with
// WithOrganizationID. It requires sub-account management permissions for the
// organization. The sub-account and all of its data are removed and cannot be
// restored; deleting the organization's last sub-account also deletes the
// organization. A repeated call for the same sub-account returns 404.
// Rate limit: 10 requests per minute per organization.
func (s *SubAccountsService) Delete(ctx context.Context, subAccountID int64) (*Response, error) {
	if s.client.organizationID == 0 {
		return nil, errNoOrganizationID
	}
	path := fmt.Sprintf("/api/organizations/%d/sub_accounts/%d", s.client.organizationID, subAccountID)
	return s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, nil)
}
