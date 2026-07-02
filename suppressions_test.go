package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestSuppressions_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/suppressions", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("email"); got != "user@example.com" {
			t.Errorf("email query = %q", got)
		}
		_, _ = w.Write([]byte(`[{"id":"64d71bf3-1276-417b-86e1-8e66f138acfe","email":"user@example.com","type":"hard bounce","sending_stream":"transactional"}]`))
	})

	suppressions, _, err := client.Suppressions.List(context.Background(), &mailtrap.SuppressionListOptions{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(suppressions) != 1 || suppressions[0].ID != "64d71bf3-1276-417b-86e1-8e66f138acfe" || suppressions[0].Type != "hard bounce" {
		t.Fatalf("suppressions = %+v", suppressions)
	}
}

func TestSuppressions_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/suppressions", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"email":"user@example.com","domain_id":12345,"sending_stream":"transactional"}`)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":"abc","email":"user@example.com","type":"manual import"}}`))
	})

	suppression, _, err := client.Suppressions.Create(context.Background(), &mailtrap.CreateSuppressionRequest{
		Email:         "user@example.com",
		DomainID:      12345,
		SendingStream: "transactional",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if suppression.ID != "abc" || suppression.Type != "manual import" {
		t.Errorf("suppression = %+v", suppression)
	}
}

func TestSuppressions_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/suppressions/abc", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"abc","email":"user@example.com"}`))
	})

	suppression, _, err := client.Suppressions.Delete(context.Background(), "abc")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if suppression.ID != "abc" {
		t.Errorf("suppression = %+v", suppression)
	}
}
