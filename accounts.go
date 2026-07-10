package mailtrap

import (
	"context"
	"net/http"
)

// AccountsService lists the Mailtrap accounts the API token can access.
type AccountsService struct {
	client *Client
}

// Account is a Mailtrap account the token has access to.
type Account struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// AccessLevels holds the token's access levels for the account.
	AccessLevels []int `json:"access_levels"`
}

// List returns the accounts the token can access. An organization-level token
// returns every account in the organization.
func (s *AccountsService) List(ctx context.Context) ([]*Account, *Response, error) {
	var accounts []*Account
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/accounts", nil, nil, &accounts)
	return accounts, resp, err
}
