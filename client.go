// Package mailtrap is the official Go client library for the Mailtrap email
// delivery platform: transactional and bulk sending, the email sandbox
// (testing), and email marketing.
//
// Construct a client with an API token and optional functional options:
//
//	client, err := mailtrap.NewClient("your-api-token")
//	if err != nil {
//		// handle error
//	}
package mailtrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/mailtrap/mailtrap-go/internal/transport"
)

// Host identifies one of Mailtrap's API hosts. Each request targets a specific
// host; values can be overridden (e.g. for tests) with WithBaseURL.
type Host int

const (
	// HostGeneral is the account/sandbox management API (https://mailtrap.io).
	HostGeneral Host = iota
	// HostSend is the transactional sending API (https://send.api.mailtrap.io).
	HostSend
	// HostBulk is the bulk sending API (https://bulk.api.mailtrap.io).
	HostBulk
	// HostSandbox is the sandbox (testing) sending API (https://sandbox.api.mailtrap.io).
	HostSandbox
)

var defaultBaseURLs = map[Host]string{
	HostGeneral: "https://mailtrap.io",
	HostSend:    "https://send.api.mailtrap.io",
	HostBulk:    "https://bulk.api.mailtrap.io",
	HostSandbox: "https://sandbox.api.mailtrap.io",
}

// ErrNoAccountID is returned by account-scoped methods when the client was
// created without an account ID. Set one with WithAccountID.
var ErrNoAccountID = errors.New("mailtrap: account ID is required; set it with WithAccountID")

// Client manages communication with the Mailtrap API.
type Client struct {
	httpClient *http.Client
	baseURLs   map[Host]string
	userAgent  string

	// accountID scopes the management resources (projects, sandboxes, ...).
	accountID int64

	// Projects manages sandbox projects.
	Projects *ProjectsService
}

// Option configures a Client in NewClient.
type Option func(*Client) error

// NewClient returns a Mailtrap API client authenticated with the given token.
// The token is required; everything else is set through options.
func NewClient(token string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, errors.New("mailtrap: API token is required")
	}

	c := &Client{
		baseURLs:  cloneBaseURLs(),
		userAgent: transport.UserAgent(),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// Wrap the HTTP client's transport so every request carries auth + UA,
	// without mutating a client the caller may have passed in.
	hc := &http.Client{}
	if c.httpClient != nil {
		*hc = *c.httpClient
	}
	hc.Transport = &transport.AuthTransport{
		Token:     token,
		UserAgent: c.userAgent,
		Base:      hc.Transport,
	}
	c.httpClient = hc

	c.Projects = &ProjectsService{client: c}

	return c, nil
}

func cloneBaseURLs() map[Host]string {
	m := make(map[Host]string, len(defaultBaseURLs))
	for h, u := range defaultBaseURLs {
		m[h] = u
	}
	return m
}

// WithHTTPClient sets the underlying *http.Client. Its transport is wrapped to
// inject authentication, so a custom transport is preserved as the base.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		if httpClient == nil {
			return errors.New("mailtrap: HTTP client must not be nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithUserAgent overrides the default User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) error {
		if userAgent == "" {
			return errors.New("mailtrap: user agent must not be empty")
		}
		c.userAgent = userAgent
		return nil
	}
}

// WithAccountID sets the account ID used by management resources (projects,
// sandboxes, messages, attachments).
func WithAccountID(accountID int64) Option {
	return func(c *Client) error {
		if accountID <= 0 {
			return fmt.Errorf("mailtrap: account ID must be positive, got %d", accountID)
		}
		c.accountID = accountID
		return nil
	}
}

// WithBaseURL overrides the base URL for a host, primarily for testing against
// an httptest server.
func WithBaseURL(host Host, rawURL string) Option {
	return func(c *Client) error {
		u := strings.TrimRight(rawURL, "/")
		if u == "" {
			return errors.New("mailtrap: base URL must not be empty")
		}
		c.baseURLs[host] = u
		return nil
	}
}

// Response wraps the HTTP response with pagination metadata.
type Response struct {
	*http.Response

	// NextPage is the next page number for paginated list endpoints, or 0 when
	// there are no further pages.
	NextPage int
}

// do sends a request to host and decodes a JSON body into out (which may be
// nil). Non-2xx responses are mapped to typed errors (see errors.go).
func (c *Client) do(ctx context.Context, host Host, method, path string, query url.Values, body, out any) (*Response, error) {
	req, err := transport.BuildRequest(ctx, method, c.baseURLs[host], path, query, body)
	if err != nil {
		return nil, fmt.Errorf("mailtrap: %w", err)
	}

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mailtrap: %w", err)
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return &Response{Response: httpResp}, fmt.Errorf("mailtrap: read response: %w", err)
	}

	resp := &Response{Response: httpResp}
	if httpResp.StatusCode >= http.StatusBadRequest {
		return resp, parseError(httpResp, data)
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return resp, fmt.Errorf("mailtrap: decode response: %w", err)
		}
	}
	return resp, nil
}

// accountPath builds an account-scoped request path from a printf-style format,
// returning ErrNoAccountID when no account ID is configured, e.g.
// accountPath("/inboxes/%d/messages/%d", sandboxID, messageID).
func (c *Client) accountPath(format string, args ...any) (string, error) {
	if c.accountID <= 0 {
		return "", ErrNoAccountID
	}
	return "/api/accounts/" + strconv.FormatInt(c.accountID, 10) + fmt.Sprintf(format, args...), nil
}
