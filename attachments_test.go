package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestSandboxAttachments_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/sandboxes/7/messages/9/attachments", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("attachment_type"); got != "inline" {
			t.Errorf("attachment_type = %q, want inline", got)
		}
		_, _ = w.Write([]byte(`[{"id":3,"message_id":9,"filename":"logo.png"}]`))
	})

	attachments, _, err := client.SandboxAttachments.List(context.Background(), 7, 9, &mailtrap.AttachmentListOptions{Type: "inline"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(attachments) != 1 || attachments[0].ID != 3 || attachments[0].Filename != "logo.png" {
		t.Fatalf("attachments = %+v", attachments)
	}
}

func TestSandboxAttachments_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/sandboxes/7/messages/9/attachments/3", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":3,"message_id":9,"filename":"logo.png","content_type":"image/png"}`))
	})

	attachment, _, err := client.SandboxAttachments.Get(context.Background(), 7, 9, 3)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if attachment.ID != 3 || attachment.ContentType != "image/png" {
		t.Errorf("attachment = %+v", attachment)
	}
}
