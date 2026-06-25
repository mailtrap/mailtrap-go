package mailtrap_test

import (
	"context"
	"errors"
	"net/http"
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
