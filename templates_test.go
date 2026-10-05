package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

const templateJSON = `{
	"id": 1234,
	"uuid": "1b5f0c2e-6f44-4f6a-9c0d-0a1b2c3d4e5f",
	"name": "Welcome",
	"category": "Onboarding",
	"subject": "Welcome aboard",
	"body_html": "<h1>Welcome!</h1>",
	"body_text": "Welcome!",
	"created_at": "2026-05-01T10:15:00.000Z",
	"updated_at": "2026-05-02T09:00:00.000Z"
}`

func TestTemplates_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/templates", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("token"); got != "2" {
			t.Errorf("token = %q, want 2", got)
		}
		if got := q.Get("per_page"); got != "10" {
			t.Errorf("per_page = %q, want 10", got)
		}
		_, _ = w.Write([]byte(`{
			"data": [` + templateJSON + `],
			"pagination": {
				"token": 2,
				"prev_token": 1,
				"next_token": 3,
				"first_url": "https://mailtrap.io/api/templates?per_page=10&token=1",
				"prev_url": "https://mailtrap.io/api/templates?per_page=10&token=1",
				"current_url": "https://mailtrap.io/api/templates?per_page=10&token=2",
				"next_url": "https://mailtrap.io/api/templates?per_page=10&token=3"
			}
		}`))
	})

	list, _, err := client.Templates.List(context.Background(), &mailtrap.TemplateListOptions{Token: 2, PerPage: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Data) != 1 || list.Data[0].ID != 1234 || list.Data[0].Name != "Welcome" {
		t.Errorf("Data = %+v", list.Data)
	}
	p := list.Pagination
	if p.Token != 2 || p.PrevToken == nil || *p.PrevToken != 1 || p.NextToken == nil || *p.NextToken != 3 {
		t.Errorf("Pagination = %+v", p)
	}
	if p.FirstURL == "" || p.CurrentURL == "" || p.NextURL == nil || p.PrevURL == nil {
		t.Errorf("Pagination URLs = %+v", p)
	}
}

func TestTemplates_ListDefaults(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("query = %q, want empty", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"data": [], "pagination": {"token": 1, "prev_token": null, "next_token": null}}`))
	})

	list, _, err := client.Templates.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Data) != 0 || list.Pagination.NextToken != nil {
		t.Errorf("list = %+v", list)
	}
}

func TestTemplates_All(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") == "" {
			_, _ = w.Write([]byte(`{
				"data": [{"id": 1, "name": "First"}, {"id": 2, "name": "Second"}],
				"pagination": {"token": 1, "prev_token": null, "next_token": 2}
			}`))
			return
		}
		if got := r.URL.Query().Get("token"); got != "2" {
			t.Errorf("token = %q, want 2", got)
		}
		_, _ = w.Write([]byte(`{
			"data": [{"id": 3, "name": "Third"}],
			"pagination": {"token": 2, "prev_token": 1, "next_token": null}
		}`))
	})

	var ids []int64
	for tpl, err := range client.Templates.All(context.Background(), nil) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		ids = append(ids, tpl.ID)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Errorf("ids = %v", ids)
	}
}

func TestTemplates_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/templates/1234", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data": ` + templateJSON + `}`))
	})

	tpl, _, err := client.Templates.Get(context.Background(), 1234)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if tpl.ID != 1234 || tpl.UUID == "" || tpl.Category != "Onboarding" || tpl.BodyHTML != "<h1>Welcome!</h1>" {
		t.Errorf("template = %+v", tpl)
	}
}

func TestTemplates_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/templates", func(w http.ResponseWriter, r *http.Request) {
		// The body must be flat — no {"email_template": ...} wrapper.
		wantJSONBody(t, r, `{
			"name": "Welcome",
			"subject": "Welcome aboard",
			"category": "Onboarding",
			"body_html": "<h1>Welcome!</h1>",
			"body_text": "Welcome!"
		}`)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data": ` + templateJSON + `}`))
	})

	req := &mailtrap.CreateTemplateRequest{
		Name:     "Welcome",
		Subject:  "Welcome aboard",
		Category: "Onboarding",
		BodyHTML: "<h1>Welcome!</h1>",
		BodyText: "Welcome!",
	}
	tpl, resp, err := client.Templates.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d", resp.StatusCode)
	}
	if tpl.ID != 1234 {
		t.Errorf("template = %+v", tpl)
	}
}

func TestTemplates_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/templates/1234", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"subject": "Welcome to Mailtrap"}`)
		_, _ = w.Write([]byte(`{"data": ` + templateJSON + `}`))
	})

	tpl, _, err := client.Templates.Update(context.Background(), 1234, &mailtrap.UpdateTemplateRequest{Subject: "Welcome to Mailtrap"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if tpl.ID != 1234 {
		t.Errorf("template = %+v", tpl)
	}
}

func TestTemplates_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/templates/1234", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.Templates.Delete(context.Background(), 1234)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
