package mailtrap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Error is the base Mailtrap API error for non-2xx responses. Status-specific
// wrappers embed it (e.g. RateLimitError, ValidationError); match with errors.As.
type Error struct {
	// StatusCode is the HTTP status code.
	StatusCode int
	// Messages holds the human-readable error message(s) returned by the API.
	Messages []string
	// Body is the raw, undecoded response body.
	Body []byte
}

func (e *Error) Error() string {
	status := fmt.Sprintf("%d %s", e.StatusCode, http.StatusText(e.StatusCode))
	if len(e.Messages) == 0 {
		return "mailtrap: " + status
	}
	return "mailtrap: " + status + ": " + strings.Join(e.Messages, "; ")
}

// UnauthorizedError indicates the API token is missing or invalid (401).
type UnauthorizedError struct{ Err *Error }

func (e *UnauthorizedError) Error() string { return e.Err.Error() }

// Unwrap exposes the base *Error to errors.As/Is.
func (e *UnauthorizedError) Unwrap() error { return e.Err }

// ForbiddenError indicates the token is valid but lacks access to the
// resource (403).
type ForbiddenError struct{ Err *Error }

func (e *ForbiddenError) Error() string { return e.Err.Error() }

// Unwrap exposes the base *Error to errors.As/Is.
func (e *ForbiddenError) Unwrap() error { return e.Err }

// NotFoundError indicates the requested resource does not exist (404).
type NotFoundError struct{ Err *Error }

func (e *NotFoundError) Error() string { return e.Err.Error() }

// Unwrap exposes the base *Error to errors.As/Is.
func (e *NotFoundError) Unwrap() error { return e.Err }

// RateLimitError indicates the API rate limit was exceeded (429).
type RateLimitError struct {
	Err *Error
	// RetryAfter is the delay advised by the Retry-After header, or 0 if absent.
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string { return e.Err.Error() }

// Unwrap exposes the base *Error to errors.As/Is.
func (e *RateLimitError) Unwrap() error { return e.Err }

// ValidationError indicates request validation failed (422). Fields maps each
// invalid attribute to its messages; the API's record-level errors use "base".
type ValidationError struct {
	Err    *Error
	Fields map[string][]string
}

func (e *ValidationError) Error() string { return e.Err.Error() }

// Unwrap exposes the base *Error to errors.As/Is.
func (e *ValidationError) Unwrap() error { return e.Err }

// parseError maps an HTTP error response to a typed error. It understands both
// Mailtrap error shapes: the sending hosts return {"success":false,"errors":[…]}
// and the general host returns {"error":…} or {"errors":…} (string or object).
func parseError(resp *http.Response, body []byte) error {
	base := &Error{StatusCode: resp.StatusCode, Body: body}

	var env struct {
		Error  string          `json:"error"`
		Errors json.RawMessage `json:"errors"`
	}
	_ = json.Unmarshal(body, &env) // best effort; body may be empty or non-JSON

	var fields map[string][]string
	switch {
	case len(env.Errors) > 0:
		base.Messages, fields = decodeErrors(env.Errors)
	case env.Error != "":
		base.Messages = []string{env.Error}
	}

	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return &RateLimitError{Err: base, RetryAfter: parseRetryAfter(resp.Header)}
	case resp.StatusCode == http.StatusUnauthorized:
		return &UnauthorizedError{Err: base}
	case resp.StatusCode == http.StatusForbidden:
		return &ForbiddenError{Err: base}
	case resp.StatusCode == http.StatusNotFound:
		return &NotFoundError{Err: base}
	case resp.StatusCode == http.StatusUnprocessableEntity, fields != nil:
		if fields == nil {
			fields = map[string][]string{}
		}
		return &ValidationError{Err: base, Fields: fields}
	default:
		return base
	}
}

// decodeErrors interprets the polymorphic "errors" field: an array of strings
// (sending hosts), a single string (general 403/429), or an object mapping
// fields to messages (general 422 validation).
func decodeErrors(raw json.RawMessage) (messages []string, fields map[string][]string) {
	var arr []string
	if json.Unmarshal(raw, &arr) == nil {
		return arr, nil
	}

	var s string
	if json.Unmarshal(raw, &s) == nil {
		return []string{s}, nil
	}

	var obj map[string][]string
	if json.Unmarshal(raw, &obj) == nil {
		fields := make([]string, 0, len(obj))
		for field := range obj {
			fields = append(fields, field)
		}
		sort.Strings(fields) // deterministic message order
		for _, field := range fields {
			joined := strings.Join(obj[field], ", ")
			// "base" carries record-level errors and is shown without a prefix,
			// matching the other Mailtrap SDKs.
			if field == "base" {
				messages = append(messages, joined)
			} else {
				messages = append(messages, field+": "+joined)
			}
		}
		return messages, obj
	}

	return nil, nil
}

// parseRetryAfter reads the Retry-After header, supporting both the delay-seconds
// and HTTP-date forms.
func parseRetryAfter(h http.Header) time.Duration {
	v := h.Get("Retry-After")
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}
