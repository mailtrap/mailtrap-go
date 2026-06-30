package mailtrap_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func sampleEmail() *mailtrap.SendRequest {
	return &mailtrap.SendRequest{
		From:    mailtrap.Address{Email: "from@example.com", Name: "Sender"},
		To:      []mailtrap.Address{{Email: "to@example.com"}},
		Subject: "Hello",
		Text:    "Hello, world",
	}
}

func TestSend_sandboxRouting(t *testing.T) {
	mux, client := setup(t, mailtrap.WithSandbox(true), mailtrap.WithSandboxID(99))
	mux.HandleFunc("POST /api/send/99", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"from":{"email":"from@example.com","name":"Sender"},"to":[{"email":"to@example.com"}],"subject":"Hello","text":"Hello, world"}`)
		_, _ = w.Write([]byte(`{"success":true,"message_ids":["abc-123"]}`))
	})

	resp, _, err := client.Send(context.Background(), sampleEmail())
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !resp.Success || len(resp.MessageIDs) != 1 || resp.MessageIDs[0] != "abc-123" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestSend_transactionalRouting(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/send", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"message_ids":["xyz"]}`))
	})

	resp, _, err := client.Send(context.Background(), sampleEmail())
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !resp.Success {
		t.Errorf("resp = %+v", resp)
	}
}

func TestSendBatch_sandboxRouting(t *testing.T) {
	mux, client := setup(t, mailtrap.WithSandbox(true), mailtrap.WithSandboxID(99))
	mux.HandleFunc("POST /api/batch/99", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"responses":[{"success":true,"message_ids":["m1"]}]}`))
	})

	req := &mailtrap.BatchSendRequest{
		Base:     &mailtrap.SendRequest{From: mailtrap.Address{Email: "from@example.com"}, Subject: "Hi", Text: "yo"},
		Requests: []mailtrap.SendRequest{{To: []mailtrap.Address{{Email: "to@example.com"}}}},
	}
	resp, _, err := client.SendBatch(context.Background(), req)
	if err != nil {
		t.Fatalf("SendBatch: %v", err)
	}
	if !resp.Success || len(resp.Responses) != 1 || resp.Responses[0].MessageIDs[0] != "m1" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestSend_errorDecodesSendShape(t *testing.T) {
	mux, client := setup(t, mailtrap.WithSandbox(true), mailtrap.WithSandboxID(99))
	mux.HandleFunc("POST /api/send/99", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"errors":["'from' is invalid"]}`))
	})

	_, _, err := client.Send(context.Background(), sampleEmail())
	var apiErr *mailtrap.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As(*Error) = false for %T", err)
	}
	if len(apiErr.Messages) != 1 || apiErr.Messages[0] != "'from' is invalid" {
		t.Errorf("messages = %v", apiErr.Messages)
	}
}

func TestSend_nilRequest(t *testing.T) {
	_, client := setup(t)
	if _, _, err := client.Send(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil request")
	}
}

// TestSend_bulkRouting proves bulk mode targets the bulk host: the transactional
// host is pointed at a server that fails the test if it is ever reached.
func TestSend_bulkRouting(t *testing.T) {
	unexpected := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("transactional host must not be called in bulk mode")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(unexpected.Close)

	mux, client := setup(t,
		mailtrap.WithBulk(true),
		mailtrap.WithBaseURL(mailtrap.HostSend, unexpected.URL),
	)
	mux.HandleFunc("POST /api/send", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"message_ids":["bulk-1"]}`))
	})

	resp, _, err := client.Send(context.Background(), sampleEmail())
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !resp.Success || len(resp.MessageIDs) != 1 || resp.MessageIDs[0] != "bulk-1" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestSendBatch_transactionalRouting(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/batch", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"responses":[{"success":true,"message_ids":["b1"]}]}`))
	})

	req := &mailtrap.BatchSendRequest{
		Base:     &mailtrap.SendRequest{From: mailtrap.Address{Email: "from@example.com"}, Subject: "Hi", Text: "yo"},
		Requests: []mailtrap.SendRequest{{To: []mailtrap.Address{{Email: "to@example.com"}}}},
	}
	resp, _, err := client.SendBatch(context.Background(), req)
	if err != nil {
		t.Fatalf("SendBatch: %v", err)
	}
	if !resp.Success || len(resp.Responses) != 1 || resp.Responses[0].MessageIDs[0] != "b1" {
		t.Errorf("resp = %+v", resp)
	}
}

// TestSend_fullMailModel checks every field of the mail model is serialized.
func TestSend_fullMailModel(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/send", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{
			"from":{"email":"from@example.com","name":"Sender"},
			"to":[{"email":"to@example.com"}],
			"cc":[{"email":"cc@example.com"}],
			"bcc":[{"email":"bcc@example.com"}],
			"reply_to":{"email":"reply@example.com"},
			"subject":"Full model",
			"text":"text body",
			"html":"<p>html body</p>",
			"category":"promo",
			"attachments":[{"content":"Zm9v","type":"text/plain","filename":"a.txt"}],
			"headers":{"X-Custom":"1"},
			"custom_variables":{"uid":"42"}
		}`)
		_, _ = w.Write([]byte(`{"success":true,"message_ids":["x"]}`))
	})

	req := &mailtrap.SendRequest{
		From:            mailtrap.Address{Email: "from@example.com", Name: "Sender"},
		To:              []mailtrap.Address{{Email: "to@example.com"}},
		Cc:              []mailtrap.Address{{Email: "cc@example.com"}},
		Bcc:             []mailtrap.Address{{Email: "bcc@example.com"}},
		ReplyTo:         &mailtrap.Address{Email: "reply@example.com"},
		Subject:         "Full model",
		Text:            "text body",
		HTML:            "<p>html body</p>",
		Category:        "promo",
		Attachments:     []mailtrap.Attachment{{Content: "Zm9v", Type: "text/plain", Filename: "a.txt"}},
		Headers:         map[string]string{"X-Custom": "1"},
		CustomVariables: map[string]any{"uid": "42"},
	}
	if _, _, err := client.Send(context.Background(), req); err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestSend_templateBody(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/send", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{
			"from":{"email":"from@example.com"},
			"to":[{"email":"to@example.com"}],
			"template_uuid":"b81aabcd-1a1e-41cf-91b6-eca0254b3d96",
			"template_variables":{"name":"Jane"}
		}`)
		_, _ = w.Write([]byte(`{"success":true,"message_ids":["x"]}`))
	})

	req := &mailtrap.SendRequest{
		From:              mailtrap.Address{Email: "from@example.com"},
		To:                []mailtrap.Address{{Email: "to@example.com"}},
		TemplateUUID:      "b81aabcd-1a1e-41cf-91b6-eca0254b3d96",
		TemplateVariables: map[string]any{"name": "Jane"},
	}
	if _, _, err := client.Send(context.Background(), req); err != nil {
		t.Fatalf("Send: %v", err)
	}
}
