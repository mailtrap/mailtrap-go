package mailtrap_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestAccountAccesses_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/account_accesses", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":42,"specifier_type":"User","specifier":{"id":1,"email":"a@b.co"},"resources":[{"resource_id":10,"resource_type":"account","access_level":100}],"permissions":{"can_read":true,"can_destroy":true}}]`))
	})

	accesses, _, err := client.AccountAccesses.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(accesses) != 1 {
		t.Fatalf("accesses = %+v", accesses)
	}
	a := accesses[0]
	if a.ID != 42 || a.SpecifierType != mailtrap.SpecifierTypeUser || a.Specifier.Email != "a@b.co" {
		t.Errorf("access = %+v (specifier %+v)", a, a.Specifier)
	}
	if a.Resources[0].AccessLevel != mailtrap.AccessLevelAdmin || !a.Permissions.CanDestroy {
		t.Errorf("resources/permissions = %+v / %+v", a.Resources[0], a.Permissions)
	}
}

func TestAccountAccesses_List_filters(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/account_accesses", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q["project_ids[]"]; !reflect.DeepEqual(got, []string{"3938"}) {
			t.Errorf("project_ids[] = %v", got)
		}
		if got := q["sandbox_ids[]"]; !reflect.DeepEqual(got, []string{"3757", "3758"}) {
			t.Errorf("sandbox_ids[] = %v", got)
		}
		if got := q["domain_ids[]"]; !reflect.DeepEqual(got, []string{"3883"}) {
			t.Errorf("domain_ids[] = %v", got)
		}
		_, _ = w.Write([]byte(`[]`))
	})

	_, _, err := client.AccountAccesses.List(context.Background(), &mailtrap.AccountAccessListOptions{
		ProjectIDs: []int64{3938},
		SandboxIDs: []int64{3757, 3758},
		DomainIDs:  []int64{3883},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestAccountAccesses_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/account_accesses/2981", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":2981}`))
	})

	resp, err := client.AccountAccesses.Delete(context.Background(), 2981)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
