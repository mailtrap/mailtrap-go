package transport

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAuthTransport_setsHeaders(t *testing.T) {
	var seen *http.Request
	at := &AuthTransport{
		Token:     "tok",
		UserAgent: "mailtrap-go/test",
		Base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			seen = r
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil
		}),
	}

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/x", nil)
	resp, err := at.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	defer resp.Body.Close()

	if got := seen.Header.Get("Authorization"); got != "Bearer tok" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer tok")
	}
	if got := seen.Header.Get("User-Agent"); got != "mailtrap-go/test" {
		t.Errorf("User-Agent = %q, want %q", got, "mailtrap-go/test")
	}
	// The RoundTripper contract forbids mutating the caller's request.
	if got := req.Header.Get("Authorization"); got != "" {
		t.Errorf("original request mutated: Authorization = %q", got)
	}
}

func TestBuildRequest(t *testing.T) {
	req, err := BuildRequest(context.Background(), http.MethodPost, "https://api.example.com", "/api/accounts/1/projects", url.Values{"page": {"2"}}, map[string]string{"name": "QA"})
	if err != nil {
		t.Fatalf("BuildRequest: %v", err)
	}
	if got := req.URL.String(); got != "https://api.example.com/api/accounts/1/projects?page=2" {
		t.Errorf("URL = %q", got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q", got)
	}
	data, _ := io.ReadAll(req.Body)
	if !strings.Contains(string(data), `"name":"QA"`) {
		t.Errorf("body = %q", data)
	}
}

func TestUserAgent(t *testing.T) {
	if ua := UserAgent(); !strings.HasPrefix(ua, "mailtrap-go/") {
		t.Errorf("UserAgent = %q, want mailtrap-go/ prefix", ua)
	}
}
