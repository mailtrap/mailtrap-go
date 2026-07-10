package mailtrap_test

import (
	"context"
	"net/http"
	"testing"
)

func TestBilling_Usage(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/billing/usage", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"billing":{"cycle_start":"2024-02-15T21:11:59.624Z","cycle_end":"2024-03-15T21:11:59.624Z"},"testing":{"plan":{"name":"Individual"},"usage":{"sent_messages_count":{"current":1234,"limit":5000},"forwarded_messages_count":{"current":0,"limit":100}}},"sending":{"plan":{"name":"Basic 10K"},"usage":{"sent_messages_count":{"current":6789,"limit":10000}}}}`))
	})

	usage, _, err := client.Billing.Usage(context.Background())
	if err != nil {
		t.Fatalf("Usage: %v", err)
	}
	if usage.Testing.Plan.Name != "Individual" || usage.Testing.Usage.SentMessagesCount.Current != 1234 {
		t.Errorf("testing = %+v", usage.Testing)
	}
	if usage.Testing.Usage.ForwardedMessagesCount.Limit != 100 {
		t.Errorf("forwarded = %+v", usage.Testing.Usage.ForwardedMessagesCount)
	}
	if usage.Sending.Usage.SentMessagesCount.Limit != 10000 {
		t.Errorf("sending = %+v", usage.Sending)
	}
	if usage.Marketing != nil {
		t.Errorf("marketing = %+v, want nil", usage.Marketing)
	}
}
