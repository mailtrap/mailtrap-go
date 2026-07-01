package mailtrap_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

// setup starts a test server routing through a ServeMux and returns the mux and
// a client pointed at it. Tests register the exact route they expect, e.g.
// mux.HandleFunc("GET /api/projects", ...), so a wrong method or path fails the
// request naturally.
func setup(t *testing.T, opts ...mailtrap.Option) (*http.ServeMux, *mailtrap.Client) {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	base := []mailtrap.Option{
		mailtrap.WithBaseURL(mailtrap.HostGeneral, srv.URL),
		mailtrap.WithBaseURL(mailtrap.HostSandbox, srv.URL),
		mailtrap.WithBaseURL(mailtrap.HostSend, srv.URL),
		mailtrap.WithBaseURL(mailtrap.HostBulk, srv.URL),
	}
	client, err := mailtrap.NewClient("test-token", append(base, opts...)...)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return mux, client
}

// wantJSONBody fails the test unless r's body matches want structurally.
func wantJSONBody(t *testing.T, r *http.Request, want string) {
	t.Helper()
	var got, exp any
	if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if err := json.Unmarshal([]byte(want), &exp); err != nil {
		t.Fatalf("bad want JSON %q: %v", want, err)
	}
	if !reflect.DeepEqual(got, exp) {
		t.Errorf("request body = %v, want %v", got, exp)
	}
}

func TestNewClient_validation(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		opts    []mailtrap.Option
		wantErr bool
	}{
		{name: "token required", token: "", wantErr: true},
		{name: "token only", token: "tok"},
		{name: "nil HTTP client", token: "tok", opts: []mailtrap.Option{mailtrap.WithHTTPClient(nil)}, wantErr: true},
		{name: "empty user agent", token: "tok", opts: []mailtrap.Option{mailtrap.WithUserAgent("")}, wantErr: true},
		{name: "empty base url", token: "tok", opts: []mailtrap.Option{mailtrap.WithBaseURL(mailtrap.HostGeneral, "")}, wantErr: true},
		{name: "sandbox requires id", token: "tok", opts: []mailtrap.Option{mailtrap.WithSandbox(true)}, wantErr: true},
		{name: "sandbox with id", token: "tok", opts: []mailtrap.Option{mailtrap.WithSandbox(true), mailtrap.WithSandboxID(1)}},
		{name: "bulk and sandbox conflict", token: "tok", opts: []mailtrap.Option{mailtrap.WithSandbox(true), mailtrap.WithSandboxID(1), mailtrap.WithBulk(true)}, wantErr: true},
		{name: "stray sandbox id ignored without sandbox", token: "tok", opts: []mailtrap.Option{mailtrap.WithSandboxID(1)}},
		{name: "negative sandbox id", token: "tok", opts: []mailtrap.Option{mailtrap.WithSandboxID(-1)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mailtrap.NewClient(tt.token, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewClient err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
