// Package transport holds the Mailtrap SDK's HTTP plumbing: token
// authentication, User-Agent injection, request building, and execution.
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"runtime/debug"
)

const (
	userAgentName = "mailtrap-go"
	modulePath    = "github.com/mailtrap/mailtrap-go"
)

// AuthTransport authenticates every request with a Bearer token and sets the
// SDK User-Agent, then delegates to Base (or http.DefaultTransport when nil).
type AuthTransport struct {
	Token     string
	UserAgent string
	Base      http.RoundTripper
}

// RoundTrip implements http.RoundTripper without mutating the caller's request.
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.Token)
	if t.UserAgent != "" && clone.Header.Get("User-Agent") == "" {
		clone.Header.Set("User-Agent", t.UserAgent)
	}

	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}

// BuildRequest constructs a JSON API request. baseURL is the host root; path is
// an absolute API path; a non-nil body is JSON-encoded (HTML escaping off).
func BuildRequest(ctx context.Context, method, baseURL, path string, query url.Values, body any) (*http.Request, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL %q: %w", baseURL, err)
	}
	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("parse request path %q: %w", path, err)
	}
	u := base.ResolveReference(rel)
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	var buf io.Reader
	if body != nil {
		b := &bytes.Buffer{}
		enc := json.NewEncoder(b)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		buf = b
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// UserAgent returns the SDK User-Agent, e.g.
// "mailtrap-go/v0.1.0 (go1.23; darwin/arm64)".
func UserAgent() string {
	return fmt.Sprintf("%s/%s (%s; %s/%s)", userAgentName, version(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// version resolves the mailtrap-go module version from build info, falling back
// to "dev" for local (untagged) builds.
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if info.Main.Path == modulePath && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	for _, dep := range info.Deps {
		if dep.Path == modulePath && dep.Version != "" {
			return dep.Version
		}
	}
	return "dev"
}
