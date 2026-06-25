package mailtrap_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mailtrap/mailtrap-go"
)

func TestErrors_typedDecoding(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		retry     string
		body      string
		check     func(t *testing.T, err error)
		wantInMsg string
	}{
		{
			name:      "general 401 uses error key",
			status:    http.StatusUnauthorized,
			body:      `{"error":"Incorrect API token"}`,
			wantInMsg: "Incorrect API token",
			check: func(t *testing.T, err error) {
				var ae *mailtrap.AuthenticationError
				if !errors.As(err, &ae) {
					t.Fatalf("errors.As(*AuthenticationError) = false for %T", err)
				}
			},
		},
		{
			name:      "general 403 uses errors string",
			status:    http.StatusForbidden,
			body:      `{"errors":"Account access forbidden"}`,
			wantInMsg: "Account access forbidden",
			check: func(t *testing.T, err error) {
				var ae *mailtrap.AuthenticationError
				if !errors.As(err, &ae) {
					t.Fatalf("errors.As(*AuthenticationError) = false for %T", err)
				}
			},
		},
		{
			name:   "422 validation object",
			status: http.StatusUnprocessableEntity,
			body:   `{"errors":{"email":["can't be blank"],"base":["invalid record"]}}`,
			check: func(t *testing.T, err error) {
				var ve *mailtrap.ValidationError
				if !errors.As(err, &ve) {
					t.Fatalf("errors.As(*ValidationError) = false for %T", err)
				}
				if got := ve.Fields["email"]; len(got) != 1 || got[0] != "can't be blank" {
					t.Errorf("Fields[email] = %v", got)
				}
				if _, ok := ve.Fields["base"]; !ok {
					t.Errorf("Fields missing base: %v", ve.Fields)
				}
			},
		},
		{
			name:      "429 rate limit with Retry-After",
			status:    http.StatusTooManyRequests,
			retry:     "30",
			body:      `{"errors":"Rate limit exceeded"}`,
			wantInMsg: "Rate limit exceeded",
			check: func(t *testing.T, err error) {
				var rle *mailtrap.RateLimitError
				if !errors.As(err, &rle) {
					t.Fatalf("errors.As(*RateLimitError) = false for %T", err)
				}
				if rle.RetryAfter != 30*time.Second {
					t.Errorf("RetryAfter = %v, want 30s", rle.RetryAfter)
				}
			},
		},
		{
			name:      "send host array errors",
			status:    http.StatusBadRequest,
			body:      `{"success":false,"errors":["'from' is invalid","subject required"]}`,
			wantInMsg: "'from' is invalid",
			check: func(t *testing.T, err error) {
				var apiErr *mailtrap.Error
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
			mux, client := setup(t, mailtrap.WithAccountID(123))
			mux.HandleFunc("GET /api/accounts/123/projects", func(w http.ResponseWriter, _ *http.Request) {
				if tt.retry != "" {
					w.Header().Set("Retry-After", tt.retry)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})

			_, _, err := client.Projects.List(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			// Every typed error unwraps to the base *Error.
			var base *mailtrap.Error
			if !errors.As(err, &base) {
				t.Fatalf("errors.As(*Error) = false for %T", err)
			}
			if base.StatusCode != tt.status {
				t.Errorf("StatusCode = %d, want %d", base.StatusCode, tt.status)
			}
			if tt.wantInMsg != "" && !strings.Contains(err.Error(), tt.wantInMsg) {
				t.Errorf("Error() = %q, want it to contain %q", err.Error(), tt.wantInMsg)
			}
			tt.check(t, err)
		})
	}
}
