package mailtrap

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// newTestClient is a white-box harness: it points the client at an httptest
// server so the unexported request pipeline (do, accountPath) can be exercised
// directly, before any resource exists to call it.
func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := NewClient("test-token",
		WithBaseURL(HostGeneral, srv.URL),
		WithBaseURL(HostSandbox, srv.URL),
		WithAccountID(123),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestDo_success(t *testing.T) {
	var gotMethod, gotPath, gotRawQuery, gotAuth, gotUA string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotRawQuery = r.Method, r.URL.Path, r.URL.RawQuery
		gotAuth, gotUA = r.Header.Get("Authorization"), r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`{"id":7}`))
	}))

	var out struct {
		ID int `json:"id"`
	}
	resp, err := c.do(context.Background(), HostGeneral, http.MethodGet, "/api/x", url.Values{"page": {"2"}}, nil, &out)
	if err != nil {
		t.Fatalf("do: %v", err)
	}

	if out.ID != 7 {
		t.Errorf("decoded ID = %d, want 7", out.ID)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Response.StatusCode = %d, want 200", resp.StatusCode)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/x" || gotRawQuery != "page=2" {
		t.Errorf("request = %s %s?%s", gotMethod, gotPath, gotRawQuery)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotUA == "" {
		t.Error("User-Agent header not set")
	}
}

// TestDo_sandboxHost exercises a second host (and a request body), so do's host
// parameter is genuinely varied across the pipeline tests.
func TestDo_sandboxHost(t *testing.T) {
	var gotPath string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"success":true}`))
	}))

	var out struct {
		Success bool `json:"success"`
	}
	if _, err := c.do(context.Background(), HostSandbox, http.MethodPost, "/api/send/1", nil, map[string]string{"subject": "hi"}, &out); err != nil {
		t.Fatalf("do: %v", err)
	}
	if gotPath != "/api/send/1" || !out.Success {
		t.Errorf("path = %q, success = %v", gotPath, out.Success)
	}
}

func TestDo_errorDecoding(t *testing.T) {
	tests := []struct {
		name   string
		status int
		retry  string
		body   string
		check  func(t *testing.T, err error)
	}{
		{
			name:   "401 -> UnauthorizedError (error key)",
			status: http.StatusUnauthorized,
			body:   `{"error":"Incorrect API token"}`,
			check: func(t *testing.T, err error) {
				var ue *UnauthorizedError
				if !errors.As(err, &ue) {
					t.Fatalf("errors.As(*UnauthorizedError) = false for %T", err)
				}
			},
		},
		{
			name:   "403 -> ForbiddenError (errors string)",
			status: http.StatusForbidden,
			body:   `{"errors":"Account access forbidden"}`,
			check: func(t *testing.T, err error) {
				var fe *ForbiddenError
				if !errors.As(err, &fe) {
					t.Fatalf("errors.As(*ForbiddenError) = false for %T", err)
				}
			},
		},
		{
			name:   "422 validation object",
			status: http.StatusUnprocessableEntity,
			body:   `{"errors":{"email":["can't be blank"],"base":["invalid record"]}}`,
			check: func(t *testing.T, err error) {
				var ve *ValidationError
				if !errors.As(err, &ve) {
					t.Fatalf("errors.As(*ValidationError) = false for %T", err)
				}
				if got := ve.Fields["email"]; len(got) != 1 || got[0] != "can't be blank" {
					t.Errorf("Fields[email] = %v", got)
				}
			},
		},
		{
			name:   "429 rate limit with Retry-After",
			status: http.StatusTooManyRequests,
			retry:  "30",
			body:   `{"errors":"Rate limit exceeded"}`,
			check: func(t *testing.T, err error) {
				var rle *RateLimitError
				if !errors.As(err, &rle) {
					t.Fatalf("errors.As(*RateLimitError) = false for %T", err)
				}
				if rle.RetryAfter != 30*time.Second {
					t.Errorf("RetryAfter = %v, want 30s", rle.RetryAfter)
				}
			},
		},
		{
			name:   "send host array errors",
			status: http.StatusBadRequest,
			body:   `{"success":false,"errors":["'from' is invalid","subject required"]}`,
			check: func(t *testing.T, err error) {
				var apiErr *Error
				if !errors.As(err, &apiErr) {
					t.Fatalf("errors.As(*Error) = false for %T", err)
				}
				if len(apiErr.Messages) != 2 {
					t.Errorf("Messages = %v, want 2", apiErr.Messages)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tt.retry != "" {
					w.Header().Set("Retry-After", tt.retry)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))

			_, err := c.do(context.Background(), HostGeneral, http.MethodGet, "/api/x", nil, nil, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var base *Error
			if !errors.As(err, &base) {
				t.Fatalf("errors.As(*Error) = false for %T", err)
			}
			if base.StatusCode != tt.status {
				t.Errorf("StatusCode = %d, want %d", base.StatusCode, tt.status)
			}
			tt.check(t, err)
		})
	}
}

func TestAccountPath(t *testing.T) {
	c, err := NewClient("tok", WithAccountID(123))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	got, err := c.accountPath("/inboxes/%d/messages/%d", 7, 9)
	if err != nil {
		t.Fatalf("accountPath: %v", err)
	}
	if want := "/api/accounts/123/inboxes/7/messages/9"; got != want {
		t.Errorf("accountPath = %q, want %q", got, want)
	}

	noAccount, err := NewClient("tok")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := noAccount.accountPath("/projects"); !errors.Is(err, ErrNoAccountID) {
		t.Errorf("err = %v, want ErrNoAccountID", err)
	}
}
