package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestWebhooks_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/webhooks", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":1,"url":"https://example.com/hook","active":true,"webhook_type":"email_sending"}]}`))
	})

	webhooks, _, err := client.Webhooks.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(webhooks) != 1 || webhooks[0].ID != 1 || webhooks[0].WebhookType != "email_sending" {
		t.Fatalf("webhooks = %+v", webhooks)
	}
}

func TestWebhooks_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/webhooks/1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"id":1,"url":"https://example.com/hook","event_types":["delivery","bounce"]}}`))
	})

	webhook, _, err := client.Webhooks.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if webhook.ID != 1 || len(webhook.EventTypes) != 2 {
		t.Errorf("webhook = %+v", webhook)
	}
}

func TestWebhooks_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/webhooks", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"webhook":{"url":"https://example.com/hook","webhook_type":"email_sending","active":false,"event_types":["delivery"]}}`)
		_, _ = w.Write([]byte(`{"data":{"id":7,"url":"https://example.com/hook","signing_secret":"s3cr3t"}}`))
	})

	webhook, _, err := client.Webhooks.Create(context.Background(), &mailtrap.CreateWebhookRequest{
		URL:         "https://example.com/hook",
		WebhookType: "email_sending",
		Active:      mailtrap.Ptr(false),
		EventTypes:  []string{"delivery"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if webhook.ID != 7 || webhook.SigningSecret != "s3cr3t" {
		t.Errorf("webhook = %+v", webhook)
	}
}

func TestWebhooks_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/webhooks/7", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"webhook":{"active":true,"payload_format":"jsonlines"}}`)
		_, _ = w.Write([]byte(`{"data":{"id":7,"active":true,"payload_format":"jsonlines"}}`))
	})

	webhook, _, err := client.Webhooks.Update(context.Background(), 7, &mailtrap.UpdateWebhookRequest{
		Active:        mailtrap.Ptr(true),
		PayloadFormat: "jsonlines",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !webhook.Active || webhook.PayloadFormat != "jsonlines" {
		t.Errorf("webhook = %+v", webhook)
	}
}

func TestWebhooks_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/webhooks/7", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"id":7}}`))
	})

	webhook, _, err := client.Webhooks.Delete(context.Background(), 7)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if webhook.ID != 7 {
		t.Errorf("webhook = %+v", webhook)
	}
}
