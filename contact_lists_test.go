package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestContactLists_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/lists", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":26730,"name":"Customers"},{"id":26731,"name":"Old Contacts"}]`))
	})

	lists, _, err := client.ContactLists.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(lists) != 2 || lists[0].ID != 26730 || lists[0].Name != "Customers" {
		t.Fatalf("lists = %+v", lists)
	}
}

func TestContactLists_ListSearch(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/lists", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("search"); got != "news" {
			t.Errorf("search = %q, want news", got)
		}
		_, _ = w.Write([]byte(`[{"id":26732,"name":"Newsletter"}]`))
	})

	lists, _, err := client.ContactLists.List(context.Background(), &mailtrap.ContactListListOptions{Search: "news"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(lists) != 1 || lists[0].Name != "Newsletter" {
		t.Fatalf("lists = %+v", lists)
	}
}

func TestContactLists_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/lists/26730", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":26730,"name":"Customers"}`))
	})

	list, _, err := client.ContactLists.Get(context.Background(), 26730)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if list.Name != "Customers" {
		t.Errorf("list = %+v", list)
	}
}

func TestContactLists_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/contacts/lists", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"name":"Customers"}`)
		_, _ = w.Write([]byte(`{"id":26730,"name":"Customers"}`))
	})

	list, _, err := client.ContactLists.Create(context.Background(), "Customers")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if list.ID != 26730 {
		t.Errorf("list = %+v", list)
	}
}

func TestContactLists_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/contacts/lists/26730", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"name":"Former Customers"}`)
		_, _ = w.Write([]byte(`{"id":26730,"name":"Former Customers"}`))
	})

	list, _, err := client.ContactLists.Update(context.Background(), 26730, "Former Customers")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if list.Name != "Former Customers" {
		t.Errorf("list = %+v", list)
	}
}

func TestContactLists_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/contacts/lists/26730", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.ContactLists.Delete(context.Background(), 26730)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
