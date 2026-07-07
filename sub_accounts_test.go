package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestSubAccounts_List(t *testing.T) {
	mux, client := setup(t, mailtrap.WithOrganizationID(1001))
	mux.HandleFunc("GET /api/organizations/1001/sub_accounts", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":12345,"name":"Development Team Account"},{"id":12346,"name":"QA Team Account"}]`))
	})

	subAccounts, _, err := client.SubAccounts.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(subAccounts) != 2 || subAccounts[0].ID != 12345 || subAccounts[0].Name != "Development Team Account" {
		t.Fatalf("subAccounts = %+v", subAccounts)
	}
}

func TestSubAccounts_Create(t *testing.T) {
	mux, client := setup(t, mailtrap.WithOrganizationID(1001))
	mux.HandleFunc("POST /api/organizations/1001/sub_accounts", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"account":{"name":"New Team Account"}}`)
		_, _ = w.Write([]byte(`{"id":12347,"name":"New Team Account"}`))
	})

	subAccount, _, err := client.SubAccounts.Create(context.Background(), "New Team Account")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if subAccount.ID != 12347 || subAccount.Name != "New Team Account" {
		t.Errorf("subAccount = %+v", subAccount)
	}
}

func TestSubAccounts_requiresOrganizationID(t *testing.T) {
	_, client := setup(t) // no WithOrganizationID
	if _, _, err := client.SubAccounts.List(context.Background()); err == nil {
		t.Error("List without organization ID: want error, got nil")
	}
	if _, _, err := client.SubAccounts.Create(context.Background(), "x"); err == nil {
		t.Error("Create without organization ID: want error, got nil")
	}
}
