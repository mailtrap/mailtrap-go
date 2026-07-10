package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestContactEvents_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/contacts/018dd5e3/events", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"name":"UserLogin","params":{"user_id":101,"is_active":true}}`)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"contact_id":"018dd5e3","contact_email":"john@example.com","name":"UserLogin","params":{"user_id":101}}`))
	})

	event, _, err := client.ContactEvents.Create(context.Background(), "018dd5e3", &mailtrap.CreateContactEventRequest{
		Name:   "UserLogin",
		Params: map[string]any{"user_id": 101, "is_active": true},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if event.ContactEmail != "john@example.com" || event.Name != "UserLogin" {
		t.Errorf("event = %+v", event)
	}
}
