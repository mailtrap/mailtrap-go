package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestPermissions_Resources(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/permissions/resources", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":1774100,"name":"Account","type":"account","access_level":100,"resources":[{"id":2550469,"name":"My Project","type":"project","access_level":100,"resources":[]}]}]`))
	})

	resources, _, err := client.Permissions.Resources(context.Background())
	if err != nil {
		t.Fatalf("Resources: %v", err)
	}
	if len(resources) != 1 || resources[0].Type != mailtrap.ResourceTypeAccount {
		t.Fatalf("resources = %+v", resources)
	}
	if len(resources[0].Resources) != 1 || resources[0].Resources[0].Type != mailtrap.ResourceTypeProject {
		t.Errorf("nested = %+v", resources[0].Resources)
	}
}

func TestPermissions_BulkUpdate(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PUT /api/account_accesses/5142/permissions/bulk", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"permissions":[{"resource_id":"3281","resource_type":"account","access_level":"viewer"},{"resource_id":"3809","resource_type":"sandbox","_destroy":true}]}`)
		_, _ = w.Write([]byte(`{"message":"Permissions have been updated!"}`))
	})

	_, err := client.Permissions.BulkUpdate(context.Background(), 5142, []*mailtrap.PermissionUpdate{
		{ResourceID: "3281", ResourceType: mailtrap.ResourceTypeAccount, AccessLevel: mailtrap.PermissionLevelViewer},
		{ResourceID: "3809", ResourceType: mailtrap.ResourceTypeSandbox, Destroy: true},
	})
	if err != nil {
		t.Fatalf("BulkUpdate: %v", err)
	}
}
