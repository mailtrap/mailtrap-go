package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestContacts_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/contacts", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"contact":{"email":"john@example.com","fields":{"first_name":"John"},"list_ids":[1,2]}}`)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":"018dd5e3","email":"john@example.com","status":"subscribed","list_ids":[1,2]}}`))
	})

	contact, _, err := client.Contacts.Create(context.Background(), &mailtrap.CreateContactRequest{
		Email:   "john@example.com",
		Fields:  map[string]any{"first_name": "John"},
		ListIDs: []int64{1, 2},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if contact.ID != "018dd5e3" || contact.Status != mailtrap.ContactStatusSubscribed {
		t.Errorf("contact = %+v", contact)
	}
}

func TestContacts_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/018dd5e3", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"id":"018dd5e3","email":"john@example.com","status":"unsubscribed"}}`))
	})

	contact, _, err := client.Contacts.Get(context.Background(), "018dd5e3")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if contact.Email != "john@example.com" || contact.Status != mailtrap.ContactStatusUnsubscribed {
		t.Errorf("contact = %+v", contact)
	}
}

func TestContacts_Get_byEmail(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/john@example.com", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"id":"018dd5e3","email":"john@example.com"}}`))
	})

	contact, _, err := client.Contacts.Get(context.Background(), "john@example.com")
	if err != nil {
		t.Fatalf("Get by email: %v", err)
	}
	if contact.ID != "018dd5e3" {
		t.Errorf("contact = %+v", contact)
	}
}

func TestContacts_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/contacts/018dd5e3", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"contact":{"email":"john@example.com","list_ids_included":[3],"list_ids_excluded":[1],"unsubscribed":true}}`)
		_, _ = w.Write([]byte(`{"action":"updated","data":{"id":"018dd5e3","email":"john@example.com","status":"unsubscribed"}}`))
	})

	upsert, _, err := client.Contacts.Update(context.Background(), "018dd5e3", &mailtrap.UpdateContactRequest{
		Email:           "john@example.com",
		ListIDsIncluded: []int64{3},
		ListIDsExcluded: []int64{1},
		Unsubscribed:    mailtrap.Ptr(true),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upsert.Action != "updated" || upsert.Contact.ID != "018dd5e3" {
		t.Errorf("upsert = %+v (contact %+v)", upsert, upsert.Contact)
	}
}

func TestContacts_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/contacts/018dd5e3", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.Contacts.Delete(context.Background(), "018dd5e3")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
