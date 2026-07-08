package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

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
		wantJSONBody(t, r, `{"name":"My API Token","resources":[{"resource_type":"account","resource_id":3229,"access_level":100}]}`)
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

func TestAPITokens_Reset(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/api_tokens/12345/reset", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":12345,"name":"My API Token","token":"newtoken123"}`))
	})

	token, _, err := client.APITokens.Reset(context.Background(), 12345)
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if token.Token != "newtoken123" {
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
