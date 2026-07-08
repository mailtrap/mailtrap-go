package mailtrap

import (
	"context"
	"fmt"
	"net/http"
)

// APITokensService manages the account's API tokens.
type APITokensService struct {
	client *Client
}

// APIToken is an API token and the permissions granted to it.
type APIToken struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Last4Digits is the last four characters of the token value.
	Last4Digits string `json:"last_4_digits,omitempty"`
	// CreatedBy names the user or token that created this token.
	CreatedBy string `json:"created_by,omitempty"`
	// ExpiresAt is the RFC 3339 expiry, or empty if the token does not expire.
	ExpiresAt string                `json:"expires_at,omitempty"`
	Resources []*APITokenPermission `json:"resources,omitempty"`
	// Token is the full token value, returned only by Create and Reset. Store it
	// securely — it is never returned again.
	Token string `json:"token,omitempty"`
}

// APITokenPermission grants a token access to one resource. It appears in both
// APIToken.Resources and CreateAPITokenRequest.Resources.
type APITokenPermission struct {
	// ResourceType identifies the kind of resource.
	ResourceType string `json:"resource_type"`
	ResourceID   int64  `json:"resource_id"`
	// AccessLevel is the level to grant on the resource.
	AccessLevel int `json:"access_level"`
}

// CreateAPITokenRequest is the payload for creating an API token. Name is
// required.
type CreateAPITokenRequest struct {
	Name      string                `json:"name"`
	Resources []*APITokenPermission `json:"resources,omitempty"`
}

// List returns all API tokens visible to the current token.
func (s *APITokensService) List(ctx context.Context) ([]*APIToken, *Response, error) {
	var tokens []*APIToken
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/api_tokens", nil, nil, &tokens)
	return tokens, resp, err
}

// Get returns an API token by ID. The full token value is not included; it is
// only returned by Create and Reset.
func (s *APITokensService) Get(ctx context.Context, tokenID int64) (*APIToken, *Response, error) {
	path := fmt.Sprintf("/api/api_tokens/%d", tokenID)
	token := new(APIToken)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, token)
	return token, resp, err
}

// Create adds an API token. The returned token's Token field holds the full
// value and is only available here — store it securely.
func (s *APITokensService) Create(ctx context.Context, req *CreateAPITokenRequest) (*APIToken, *Response, error) {
	token := new(APIToken)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, "/api/api_tokens", nil, req, token)
	return token, resp, err
}

// Reset expires the token and issues a replacement with the same permissions.
// The returned token's Token field holds the new value; store it securely.
func (s *APITokensService) Reset(ctx context.Context, tokenID int64) (*APIToken, *Response, error) {
	path := fmt.Sprintf("/api/api_tokens/%d/reset", tokenID)
	token := new(APIToken)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, path, nil, nil, token)
	return token, resp, err
}

// Delete permanently removes an API token by ID.
func (s *APITokensService) Delete(ctx context.Context, tokenID int64) (*Response, error) {
	path := fmt.Sprintf("/api/api_tokens/%d", tokenID)
	return s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, nil)
}
