package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestContactImports_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/contacts/imports", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"contacts":[{"email":"a@example.com","fields":{"first_name":"A"},"list_ids_included":[1],"list_ids_excluded":[4]}]}`)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"status":"started"}`))
	})

	imp, _, err := client.ContactImports.Create(context.Background(), []*mailtrap.ImportContact{
		{Email: "a@example.com", Fields: map[string]any{"first_name": "A"}, ListIDsIncluded: []int64{1}, ListIDsExcluded: []int64{4}},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if imp.ID != 1 || imp.Status != mailtrap.ContactImportStarted {
		t.Errorf("import = %+v", imp)
	}
}

func TestContactImports_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/imports/1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":1,"status":"finished","created_contacts_count":10,"updated_contacts_count":2,"contacts_over_limit_count":0}`))
	})

	imp, _, err := client.ContactImports.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if imp.Status != mailtrap.ContactImportFinished || imp.CreatedContactsCount != 10 || imp.UpdatedContactsCount != 2 {
		t.Errorf("import = %+v", imp)
	}
}
