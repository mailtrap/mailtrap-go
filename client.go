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
	"errors"
	"net/http"

	"github.com/mailtrap/mailtrap-go/internal/transport"
)

// Client manages communication with the Mailtrap API.
type Client struct {
	httpClient *http.Client
	userAgent  string
}

// Option configures a Client in NewClient.
type Option func(*Client) error

// NewClient returns a Mailtrap API client authenticated with the given token.
// The token is required; everything else is set through options.
func NewClient(token string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, errors.New("mailtrap: API token is required")
	}

	c := &Client{userAgent: transport.UserAgent()}
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

	return c, nil
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
