package mailtrap_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

// wantRawBody fails the test unless r's body is exactly want, proving whether
// the expires_at key is absent, null, or a string on the wire.
func wantRawBody(t *testing.T, r *http.Request, want string) {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	if got := strings.TrimSpace(string(b)); got != want {
		t.Errorf("request body = %q, want %q", got, want)
	}
}

func TestAPITokens_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/api_tokens", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":12345,"name":"My API Token","last_4_digits":"x7k9","resources":[{"resource_type":"account","resource_id":3229,"access_level":100}]}]`))
	})

	tokens, _, err := client.APITokens.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tokens) != 1 || tokens[0].ID != 12345 || tokens[0].Last4Digits != "x7k9" {
		t.Fatalf("tokens = %+v", tokens)
	}
	if r := tokens[0].Resources[0]; r.ResourceID != 3229 || r.AccessLevel != mailtrap.AccessLevelAdmin {
		t.Errorf("resource = %+v", r)
	}
}

func TestAPITokens_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/api_tokens/12345", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":12345,"name":"My API Token","last_4_digits":"x7k9"}`))
	})

	token, _, err := client.APITokens.Get(context.Background(), 12345)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if token.ID != 12345 || token.Token != "" {
		t.Errorf("token = %+v", token)
	}
}

func TestAPITokens_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/api_tokens", func(w http.ResponseWriter, r *http.Request) {
		wantRawBody(t, r, `{"name":"My API Token","resources":[{"resource_type":"account","resource_id":3229,"access_level":100}]}`)
		_, _ = w.Write([]byte(`{"id":12345,"name":"My API Token","token":"a1b2c3d4e5f6"}`))
	})

	token, _, err := client.APITokens.Create(context.Background(), &mailtrap.CreateAPITokenRequest{
		Name: "My API Token",
		Resources: []*mailtrap.APITokenPermission{
			{ResourceType: mailtrap.ResourceTypeAccount, ResourceID: 3229, AccessLevel: mailtrap.AccessLevelAdmin},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if token.ID != 12345 || token.Token != "a1b2c3d4e5f6" {
		t.Errorf("token = %+v", token)
	}
}

func TestAPITokens_Create_expiresAt(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/api_tokens", func(w http.ResponseWriter, r *http.Request) {
		wantRawBody(t, r, `{"name":"My API Token","expires_at":"2027-06-01T00:00:00Z"}`)
		_, _ = w.Write([]byte(`{"id":12345,"name":"My API Token","expires_at":"2027-06-01T00:00:00Z","token":"a1b2c3d4e5f6"}`))
	})

	token, _, err := client.APITokens.Create(context.Background(), &mailtrap.CreateAPITokenRequest{
		Name:      "My API Token",
		ExpiresAt: mailtrap.ExpiresAt("2027-06-01T00:00:00Z"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if token.ExpiresAt != "2027-06-01T00:00:00Z" {
		t.Errorf("token = %+v", token)
	}
}

func TestAPITokens_Create_neverExpires(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/api_tokens", func(w http.ResponseWriter, r *http.Request) {
		wantRawBody(t, r, `{"name":"My API Token","expires_at":null}`)
		_, _ = w.Write([]byte(`{"id":12345,"name":"My API Token","expires_at":null,"token":"a1b2c3d4e5f6"}`))
	})

	token, _, err := client.APITokens.Create(context.Background(), &mailtrap.CreateAPITokenRequest{
		Name:      "My API Token",
		ExpiresAt: mailtrap.NeverExpires(),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if token.ExpiresAt != "" {
		t.Errorf("token = %+v", token)
	}
}

func TestAPITokens_Create_expirationRejected(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/api_tokens", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"errors":{"expires_at":["must be in the future"]}}`))
	})

	_, _, err := client.APITokens.Create(context.Background(), &mailtrap.CreateAPITokenRequest{
		Name:      "My API Token",
		ExpiresAt: mailtrap.ExpiresAt("2020-01-01T00:00:00Z"),
	})
	var ve *mailtrap.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("errors.As(*ValidationError) = false for %T", err)
	}
	if got := ve.Fields["expires_at"]; len(got) != 1 || got[0] != "must be in the future" {
		t.Errorf("Fields[expires_at] = %v", got)
	}
}

func TestCreateAPITokenRequest_marshalExpiresAt(t *testing.T) {
	tests := []struct {
		name string
		req  *mailtrap.CreateAPITokenRequest
		want string
	}{
		{
			name: "nil omits the key",
			req:  &mailtrap.CreateAPITokenRequest{Name: "t"},
			want: `{"name":"t"}`,
		},
		{
			name: "NeverExpires writes explicit null",
			req:  &mailtrap.CreateAPITokenRequest{Name: "t", ExpiresAt: mailtrap.NeverExpires()},
			want: `{"name":"t","expires_at":null}`,
		},
		{
			name: "ExpiresAt writes the date-time",
			req:  &mailtrap.CreateAPITokenRequest{Name: "t", ExpiresAt: mailtrap.ExpiresAt("2027-06-01T00:00:00Z")},
			want: `{"name":"t","expires_at":"2027-06-01T00:00:00Z"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAPITokens_Reset(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/api_tokens/12345/reset", func(w http.ResponseWriter, r *http.Request) {
		wantRawBody(t, r, "")
		if ct := r.Header.Get("Content-Type"); ct != "" {
			t.Errorf("Content-Type = %q, want empty", ct)
		}
		_, _ = w.Write([]byte(`{"id":12345,"name":"My API Token","token":"newtoken123"}`))
	})

	token, _, err := client.APITokens.Reset(context.Background(), 12345, nil)
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if token.Token != "newtoken123" {
		t.Errorf("token = %+v", token)
	}
}

func TestAPITokens_Reset_expiresAt(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/api_tokens/12345/reset", func(w http.ResponseWriter, r *http.Request) {
		wantRawBody(t, r, `{"expires_at":"2027-06-01T00:00:00Z"}`)
		_, _ = w.Write([]byte(`{"id":12345,"name":"My API Token","expires_at":"2027-06-01T00:00:00Z","token":"newtoken123"}`))
	})

	token, _, err := client.APITokens.Reset(context.Background(), 12345, &mailtrap.ResetAPITokenRequest{
		ExpiresAt: mailtrap.ExpiresAt("2027-06-01T00:00:00Z"),
	})
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if token.ExpiresAt != "2027-06-01T00:00:00Z" {
		t.Errorf("token = %+v", token)
	}
}

func TestAPITokens_Reset_neverExpires(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/api_tokens/12345/reset", func(w http.ResponseWriter, r *http.Request) {
		wantRawBody(t, r, `{"expires_at":null}`)
		_, _ = w.Write([]byte(`{"id":12345,"name":"My API Token","expires_at":null,"token":"newtoken123"}`))
	})

	token, _, err := client.APITokens.Reset(context.Background(), 12345, &mailtrap.ResetAPITokenRequest{
		ExpiresAt: mailtrap.NeverExpires(),
	})
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if token.ExpiresAt != "" {
		t.Errorf("token = %+v", token)
	}
}

func TestAPITokens_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/api_tokens/12345", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.APITokens.Delete(context.Background(), 12345)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
