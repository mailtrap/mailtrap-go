package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestContactExports_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/contacts/exports", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"filters":[{"name":"list_id","operator":"equal","value":[101,102]},{"name":"subscription_status","operator":"equal","value":"subscribed"}]}`)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"status":"started","url":null}`))
	})

	export, _, err := client.ContactExports.Create(context.Background(), []*mailtrap.ContactExportFilter{
		{Name: mailtrap.ContactExportFilterListID, Operator: mailtrap.ContactExportOperatorEqual, Value: []int64{101, 102}},
		{Name: mailtrap.ContactExportFilterSubscriptionStatus, Operator: mailtrap.ContactExportOperatorEqual, Value: mailtrap.ContactStatusSubscribed},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if export.ID != 1 || export.URL != nil {
		t.Errorf("export = %+v", export)
	}
}

func TestContactExports_Create_noFilters(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/contacts/exports", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{}`)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":2,"status":"created"}`))
	})

	export, _, err := client.ContactExports.Create(context.Background(), nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if export.ID != 2 {
		t.Errorf("export = %+v", export)
	}
}

func TestContactExports_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/exports/1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":1,"status":"finished","url":"https://example.com/export.csv"}`))
	})

	export, _, err := client.ContactExports.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if export.Status != mailtrap.ContactExportFinished || export.URL == nil || *export.URL != "https://example.com/export.csv" {
		t.Errorf("export = %+v", export)
	}
}
