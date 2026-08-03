package mailtrap

import (
	"context"
	"encoding/json"
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

// TokenExpiration is an optional token expiration as an RFC 3339 date-time.
// Leave the request field nil for the server default (a 1-year default is
// being rolled out). Use NeverExpires for a token that never expires. Past or
// more-than-5-years-ahead values are rejected with 422.
type TokenExpiration struct {
	value string
	never bool
}

// ExpiresAt returns a token expiration at the given RFC 3339 date-time, e.g.
// "2027-06-01T00:00:00Z".
func ExpiresAt(rfc3339 string) *TokenExpiration {
	return &TokenExpiration{value: rfc3339}
}

// NeverExpires returns a token expiration for a token that never expires. It
// serializes as an explicit "expires_at": null.
func NeverExpires() *TokenExpiration {
	return &TokenExpiration{never: true}
}

// MarshalJSON encodes the RFC 3339 date-time, or null for NeverExpires.
func (e TokenExpiration) MarshalJSON() ([]byte, error) {
	if e.never {
		return []byte("null"), nil
	}
	return json.Marshal(e.value)
}

// CreateAPITokenRequest is the payload for creating an API token. Name is
// required.
type CreateAPITokenRequest struct {
	Name string `json:"name"`
	// ExpiresAt is the optional token expiration. Nil omits the field and
	// applies the server default; see TokenExpiration.
	ExpiresAt *TokenExpiration      `json:"expires_at,omitempty"`
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
