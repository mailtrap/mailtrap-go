package mailtrap_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

// jsonMessages builds a JSON array of n messages with sequential IDs from start.
func jsonMessages(start, n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = fmt.Sprintf(`{"id":%d}`, start+i)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func TestSandboxMessages_ListPaginationAndAll(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/sandboxes/7/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "", "1":
			_, _ = w.Write([]byte(jsonMessages(1, 30))) // full page → more pages
		case "2":
			_, _ = w.Write([]byte(jsonMessages(31, 5))) // partial page → last
		default:
			_, _ = w.Write([]byte("[]"))
		}
	})

	msgs, resp, err := client.SandboxMessages.List(context.Background(), 7, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(msgs) != 30 {
		t.Fatalf("len(msgs) = %d, want 30", len(msgs))
	}
	if resp.NextPage != 2 {
		t.Errorf("NextPage = %d, want 2", resp.NextPage)
	}

	var count int
	for msg, err := range client.SandboxMessages.All(context.Background(), 7, nil) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		if msg == nil {
			t.Fatal("All yielded nil message")
		}
		count++
	}
	if count != 35 {
		t.Errorf("All yielded %d messages, want 35", count)
	}
}

func TestSandboxMessages_ListSearch(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/sandboxes/7/messages", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("search"); got != "welcome" {
			t.Errorf("search = %q, want welcome", got)
		}
		_, _ = w.Write([]byte("[]"))
	})

	if _, _, err := client.SandboxMessages.List(context.Background(), 7, &mailtrap.MessageListOptions{Search: "welcome"}); err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestSandboxMessages_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/sandboxes/7/messages/9", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"message":{"is_read":true}}`)
		_, _ = w.Write([]byte(`{"id":9,"is_read":true}`))
	})

	msg, _, err := client.SandboxMessages.Update(context.Background(), 7, 9, true)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !msg.IsRead {
		t.Errorf("IsRead = false, want true")
	}
}

func TestSandboxMessages_Forward(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/sandboxes/7/messages/9/forward", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"email":"qa@example.com"}`)
		_, _ = w.Write([]byte(`{"message":"Your email message was successfully forwarded."}`))
	})

	if _, err := client.SandboxMessages.Forward(context.Background(), 7, 9, "qa@example.com"); err != nil {
		t.Fatalf("Forward: %v", err)
	}
}

func TestSandboxMessages_SpamReport(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/sandboxes/7/messages/9/spam_report", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"report":{"Score":1.2,"Spam":false,"Threshold":5,"Details":[{"Pts":0.5,"RuleName":"HTML_MESSAGE"}]}}`))
	})

	report, _, err := client.SandboxMessages.SpamReport(context.Background(), 7, 9)
	if err != nil {
		t.Fatalf("SpamReport: %v", err)
	}
	if report.Score != 1.2 || report.Spam {
		t.Errorf("report = %+v", report)
	}
	if len(report.Details) != 1 || report.Details[0].Pts != 0.5 {
		t.Errorf("details = %+v", report.Details)
	}
}

func TestSandboxMessages_Headers(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/sandboxes/7/messages/9/mail_headers", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"headers":{"subject":"Hi","from":"a@x.io"}}`))
	})

	headers, _, err := client.SandboxMessages.Headers(context.Background(), 7, 9)
	if err != nil {
		t.Fatalf("Headers: %v", err)
	}
	if headers["subject"] != "Hi" {
		t.Errorf("headers = %+v", headers)
	}
}

func TestSandboxMessages_RawBodies(t *testing.T) {
	bodies := map[string]struct {
		segment string
		call    func(*mailtrap.Client) ([]byte, error)
	}{
		"text": {"body.txt", func(c *mailtrap.Client) ([]byte, error) {
			b, _, err := c.SandboxMessages.Text(context.Background(), 7, 9)
			return b, err
		}},
		"raw": {"body.raw", func(c *mailtrap.Client) ([]byte, error) {
			b, _, err := c.SandboxMessages.Raw(context.Background(), 7, 9)
			return b, err
		}},
		"htmlsource": {"body.htmlsource", func(c *mailtrap.Client) ([]byte, error) {
			b, _, err := c.SandboxMessages.HTMLSource(context.Background(), 7, 9)
			return b, err
		}},
		"html": {"body.html", func(c *mailtrap.Client) ([]byte, error) {
			b, _, err := c.SandboxMessages.HTML(context.Background(), 7, 9)
			return b, err
		}},
		"eml": {"body.eml", func(c *mailtrap.Client) ([]byte, error) {
			b, _, err := c.SandboxMessages.EML(context.Background(), 7, 9)
			return b, err
		}},
	}
	for name, tc := range bodies {
		t.Run(name, func(t *testing.T) {
			mux, client := setup(t)
			mux.HandleFunc("GET /api/sandboxes/7/messages/9/"+tc.segment, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte("raw " + name))
			})
			got, err := tc.call(client)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			if string(got) != "raw "+name {
				t.Errorf("body = %q, want %q", got, "raw "+name)
			}
		})
	}
}
