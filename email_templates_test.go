package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestEmailTemplates_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/email_templates", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":26730,"uuid":"018dd5e3","name":"Promo","category":"Promotional"}]`))
	})

	templates, _, err := client.EmailTemplates.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(templates) != 1 || templates[0].ID != 26730 || templates[0].Name != "Promo" {
		t.Fatalf("templates = %+v", templates)
	}
}

func TestEmailTemplates_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/email_templates/26730", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":26730,"uuid":"018dd5e3","name":"Promo","body_html":"<div>Hi</div>"}`))
	})

	template, _, err := client.EmailTemplates.Get(context.Background(), 26730)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if template.ID != 26730 || template.BodyHTML != "<div>Hi</div>" {
		t.Errorf("template = %+v", template)
	}
}

func TestEmailTemplates_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/email_templates", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"email_template":{"name":"Promo","subject":"Sale","category":"Promotional","body_html":"<div>Hi</div>"}}`)
		_, _ = w.Write([]byte(`{"id":9,"name":"Promo","subject":"Sale","category":"Promotional"}`))
	})

	template, _, err := client.EmailTemplates.Create(context.Background(), &mailtrap.EmailTemplateRequest{
		Name:     "Promo",
		Subject:  "Sale",
		Category: "Promotional",
		BodyHTML: "<div>Hi</div>",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if template.ID != 9 {
		t.Errorf("template = %+v", template)
	}
}

func TestEmailTemplates_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/email_templates/9", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"email_template":{"subject":"New subject"}}`)
		_, _ = w.Write([]byte(`{"id":9,"subject":"New subject"}`))
	})

	template, _, err := client.EmailTemplates.Update(context.Background(), 9, &mailtrap.EmailTemplateRequest{Subject: "New subject"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if template.Subject != "New subject" {
		t.Errorf("template = %+v", template)
	}
}

func TestEmailTemplates_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/email_templates/9", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if _, err := client.EmailTemplates.Delete(context.Background(), 9); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
