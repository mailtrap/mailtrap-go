package mailtrap_test

import (
	"context"
	"net/http"
	"testing"
)

func TestAccounts_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/accounts", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":26730,"name":"James","access_levels":[100]},{"id":26731,"name":"John","access_levels":[1000]}]`))
	})

	accounts, _, err := client.Accounts.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("accounts = %+v", accounts)
	}
	if accounts[0].ID != 26730 || accounts[0].Name != "James" || accounts[0].AccessLevels[0] != 100 {
		t.Errorf("account[0] = %+v", accounts[0])
	}
}
