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

// Client manages communication with the Mailtrap API.
type Client struct {
	httpClient *http.Client
	baseURLs   map[Host]string
	userAgent  string

	// Sending configuration. sandbox and bulk are mutually exclusive; sandbox
	// sends require sandboxID.
	sandbox   bool
	bulk      bool
	sandboxID int64

	// Projects manages sandbox projects.
	Projects *ProjectsService
	// Sandboxes manages sandboxes (testing inboxes) and their actions.
	Sandboxes *SandboxesService
	// SandboxMessages manages messages captured by a sandbox.
	SandboxMessages *SandboxMessagesService
	// SandboxAttachments reads attachments of sandbox messages.
	SandboxAttachments *SandboxAttachmentsService

	// SendingDomains manages sending domains and their compliance settings.
	SendingDomains *SendingDomainsService
	// Suppressions manages the account's do-not-send list.
	Suppressions *SuppressionsService
}

// Ptr returns a pointer to v, for setting optional pointer request fields such
// as UpdateDomainRequest.OpenTrackingEnabled: mailtrap.Ptr(false).
func Ptr[T any](v T) *T { return &v }

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

	if err := c.validateSendConfig(); err != nil {
		return nil, err
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
	c.Sandboxes = &SandboxesService{client: c}
	c.SandboxMessages = &SandboxMessagesService{client: c}
	c.SandboxAttachments = &SandboxAttachmentsService{client: c}
	c.SendingDomains = &SendingDomainsService{client: c}
	c.Suppressions = &SuppressionsService{client: c}

	return c, nil
}

// validateSendConfig enforces the two real send-mode invariants. A stray
// sandboxID outside sandbox mode is ignored, so callers can toggle WithSandbox
// from configuration without also clearing the ID.
func (c *Client) validateSendConfig() error {
	switch {
	case c.bulk && c.sandbox:
		return errors.New("mailtrap: bulk and sandbox modes are mutually exclusive")
	case c.sandbox && c.sandboxID == 0:
		return errors.New("mailtrap: WithSandboxID is required in sandbox mode")
	default:
		return nil
	}
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

// WithSandbox routes Send/SendBatch to the sandbox host. Pair it with
// WithSandboxID. Toggle it from configuration to switch environments without
// changing call sites, e.g. WithSandbox(env != "production").
func WithSandbox(enabled bool) Option {
	return func(c *Client) error {
		c.sandbox = enabled
		return nil
	}
}

// WithBulk routes Send/SendBatch to the bulk host. Mutually exclusive with
// WithSandbox.
func WithBulk(enabled bool) Option {
	return func(c *Client) error {
		c.bulk = enabled
		return nil
	}
}

// WithSandboxID sets the sandbox that Send/SendBatch deliver to. Required in
// sandbox mode and ignored otherwise.
func WithSandboxID(sandboxID int64) Option {
	return func(c *Client) error {
		if sandboxID <= 0 {
			return fmt.Errorf("mailtrap: sandbox ID must be valid, got %d", sandboxID)
		}
		c.sandboxID = sandboxID
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

// doRaw sends a GET to the general host and returns the undecoded response
// body. Used by the sandbox message body getters, which return text or binary
// rather than JSON.
func (c *Client) doRaw(ctx context.Context, path string) ([]byte, *Response, error) {
	req, err := transport.BuildRequest(ctx, http.MethodGet, c.baseURLs[HostGeneral], path, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("mailtrap: %w", err)
	}

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("mailtrap: %w", err)
	}
	defer httpResp.Body.Close()

	data, err := io.ReadAll(httpResp.Body)
	resp := &Response{Response: httpResp}
	if err != nil {
		return nil, resp, fmt.Errorf("mailtrap: read response: %w", err)
	}
	if httpResp.StatusCode >= http.StatusBadRequest {
		return nil, resp, parseError(httpResp, data)
	}
	return data, resp, nil
}
