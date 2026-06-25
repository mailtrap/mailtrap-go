package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestSandboxes_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/sandboxes", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":7,"name":"Inbox","project_id":1}]`))
	})

	sandboxes, _, err := client.Sandboxes.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sandboxes) != 1 || sandboxes[0].ID != 7 || sandboxes[0].ProjectID != 1 {
		t.Fatalf("sandboxes = %+v", sandboxes)
	}
}

func TestSandboxes_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/projects/1/sandboxes", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"sandbox":{"name":"QA"}}`)
		_, _ = w.Write([]byte(`{"id":7,"name":"QA","project_id":1}`))
	})

	sandbox, _, err := client.Sandboxes.Create(context.Background(), 1, "QA")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if sandbox.ID != 7 || sandbox.Name != "QA" {
		t.Errorf("sandbox = %+v", sandbox)
	}
}

func TestSandboxes_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/sandboxes/7", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"sandbox":{"name":"Renamed","email_username":"qa"}}`)
		_, _ = w.Write([]byte(`{"id":7,"name":"Renamed","email_username":"qa"}`))
	})

	sandbox, _, err := client.Sandboxes.Update(context.Background(), 7, &mailtrap.SandboxUpdateRequest{
		Name:          "Renamed",
		EmailUsername: "qa",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if sandbox.Name != "Renamed" {
		t.Errorf("sandbox = %+v", sandbox)
	}
}

func TestSandboxes_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/sandboxes/7", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":7,"name":"QA"}`))
	})

	sandbox, _, err := client.Sandboxes.Delete(context.Background(), 7)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if sandbox.ID != 7 {
		t.Errorf("sandbox = %+v", sandbox)
	}
}

func TestSandboxes_Actions(t *testing.T) {
	actions := map[string]struct {
		segment string
		call    func(*mailtrap.Client) error
	}{
		"clean": {"clean", func(c *mailtrap.Client) error { _, _, err := c.Sandboxes.Clean(context.Background(), 7); return err }},
		"mark all read": {"all_read", func(c *mailtrap.Client) error {
			_, _, err := c.Sandboxes.MarkAllRead(context.Background(), 7)
			return err
		}},
		"reset credentials": {"reset_credentials", func(c *mailtrap.Client) error {
			_, _, err := c.Sandboxes.ResetCredentials(context.Background(), 7)
			return err
		}},
		"toggle email": {"toggle_email_username", func(c *mailtrap.Client) error {
			_, _, err := c.Sandboxes.ToggleEmailAddress(context.Background(), 7)
			return err
		}},
		"reset email": {"reset_email_username", func(c *mailtrap.Client) error {
			_, _, err := c.Sandboxes.ResetEmailAddress(context.Background(), 7)
			return err
		}},
	}
	for name, tc := range actions {
		t.Run(name, func(t *testing.T) {
			mux, client := setup(t)
			mux.HandleFunc("PATCH /api/sandboxes/7/"+tc.segment, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"id":7}`))
			})
			if err := tc.call(client); err != nil {
				t.Fatalf("%s: %v", name, err)
			}
		})
	}
}
