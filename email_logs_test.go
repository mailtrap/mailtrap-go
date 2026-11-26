package mailtrap_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestEmailLogs_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/email_logs", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("filters[sent_after]"); got != "2025-01-01T00:00:00Z" {
			t.Errorf("sent_after = %q", got)
		}
		if got := q.Get("filters[to][operator]"); got != "ci_equal" {
			t.Errorf("to operator = %q", got)
		}
		if got := q.Get("filters[to][value]"); got != "recipient@example.com" {
			t.Errorf("to value (scalar) = %q", got)
		}
		if got := q["filters[status][value][]"]; !reflect.DeepEqual(got, []string{"delivered", "enqueued"}) {
			t.Errorf("status value (array) = %v", got)
		}
		_, _ = w.Write([]byte(`{"messages":[{"message_id":"a1","status":"delivered","to":"recipient@example.com","sending_domain_id":3938,"template_id":100}],"total_count":150,"next_page_cursor":"a1"}`))
	})

	list, _, err := client.EmailLogs.List(context.Background(), &mailtrap.EmailLogsListOptions{
		SentAfter: "2025-01-01T00:00:00Z",
		Filters: map[string]mailtrap.LogFilter{
			"to":     {Operator: "ci_equal", Values: []string{"recipient@example.com"}},
			"status": {Operator: "equal", Values: []string{"delivered", "enqueued"}},
		},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if list.TotalCount != 150 || list.NextPageCursor != "a1" || len(list.Messages) != 1 {
		t.Fatalf("list = %+v", list)
	}
	if list.Messages[0].TemplateID == nil || *list.Messages[0].TemplateID != 100 {
		t.Errorf("template_id = %v", list.Messages[0].TemplateID)
	}
	if list.Messages[0].DomainID != 3938 {
		t.Errorf("DomainID = %d", list.Messages[0].DomainID)
	}
}

func TestEmailLogs_All(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/email_logs", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("search_after") {
		case "":
			_, _ = w.Write([]byte(`{"messages":[{"message_id":"a1"}],"total_count":2,"next_page_cursor":"a1"}`))
		case "a1":
			_, _ = w.Write([]byte(`{"messages":[{"message_id":"a2"}],"total_count":2,"next_page_cursor":null}`))
		default:
			t.Errorf("unexpected search_after %q", r.URL.Query().Get("search_after"))
		}
	})

	var ids []string
	for msg, err := range client.EmailLogs.All(context.Background(), nil) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		ids = append(ids, msg.MessageID)
	}
	if !reflect.DeepEqual(ids, []string{"a1", "a2"}) {
		t.Errorf("ids = %v", ids)
	}
}

func TestEmailLogs_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/email_logs/a1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message_id":"a1","status":"delivered","sending_domain_id":3938,"raw_message_url":"https://x/eml","events":[{"event_type":"click","created_at":"2025-01-15T10:35:00Z","details":{"click_url":"https://example.com/c"}}]}`))
	})

	msg, _, err := client.EmailLogs.Get(context.Background(), "a1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if msg.RawMessageURL != "https://x/eml" || len(msg.Events) != 1 {
		t.Fatalf("msg = %+v", msg)
	}
	if msg.DomainID != 3938 {
		t.Errorf("DomainID = %d", msg.DomainID)
	}
	if msg.Events[0].EventType != "click" || msg.Events[0].Details.ClickURL != "https://example.com/c" {
		t.Errorf("event = %+v", msg.Events[0])
	}
}

func TestEmailLogs_Get_topicEvents(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/email_logs/a1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message_id":"a1","events":[` +
			`{"event_type":"topic_unsubscribe","created_at":"2025-01-15T10:35:00Z","details":{"web_ip_address":"198.51.100.7","subscription_topic_id":12}},` +
			`{"event_type":"topic_resubscribe","created_at":"2025-01-16T10:35:00Z","details":{"web_ip_address":null,"subscription_topic_id":12}},` +
			`{"event_type":"future_event","created_at":"2025-01-17T10:35:00Z","details":{"something_new":{"nested":true},"click_url":"https://example.com/c"}}]}`))
	})

	msg, _, err := client.EmailLogs.Get(context.Background(), "a1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(msg.Events) != 3 {
		t.Fatalf("events = %+v", msg.Events)
	}
	unsub, resub, unknown := msg.Events[0], msg.Events[1], msg.Events[2]
	if unsub.EventType != mailtrap.MessageEventTypeTopicUnsubscribe || unsub.Details.WebIPAddress != "198.51.100.7" ||
		unsub.Details.SubscriptionTopicID == nil || *unsub.Details.SubscriptionTopicID != 12 {
		t.Errorf("topic_unsubscribe = %+v", unsub)
	}
	if resub.EventType != mailtrap.MessageEventTypeTopicResubscribe || resub.Details.WebIPAddress != "" ||
		resub.Details.SubscriptionTopicID == nil || *resub.Details.SubscriptionTopicID != 12 {
		t.Errorf("topic_resubscribe = %+v", resub)
	}
	// An event type this version doesn't know decodes with its raw type and
	// whatever details it shares with the known events.
	if unknown.EventType != "future_event" || unknown.Details.ClickURL != "https://example.com/c" || unknown.Details.SubscriptionTopicID != nil {
		t.Errorf("unknown event = %+v", unknown)
	}
}
