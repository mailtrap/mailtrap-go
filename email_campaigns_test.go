package mailtrap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

const campaignJSON = `{
	"id": 4567,
	"domain_id": 4321,
	"domain_name": "acme.com",
	"name": "Spring Sale",
	"from_local_part": "news",
	"from_display_name": "Acme Marketing",
	"reply_to": {"display_name": "Acme Support", "local_part": "support", "domain": "acme.com"},
	"current_state": "draft",
	"current_state_metadata": {},
	"created_at": "2026-05-01T10:15:00.000Z",
	"updated_at": "2026-05-02T09:00:00.000Z",
	"last_started_at": null,
	"recipient_total_count": null,
	"contact_list_ids": [55, 56],
	"contact_segment_ids": [12],
	"delivery_mode": "rapid",
	"delivery_options": {"emails_per_hour": null},
	"template": {"id": 789, "subject": "Spring is here", "merge_tags": ["first_name"], "body_html": "<html></html>", "body_text": null}
}`

func TestEmailCampaigns_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/email_campaigns", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("token"); got != "2" {
			t.Errorf("token = %q, want 2", got)
		}
		if got := q.Get("per_page"); got != "10" {
			t.Errorf("per_page = %q, want 10", got)
		}
		if got := q.Get("search"); got != "spring" {
			t.Errorf("search = %q, want spring", got)
		}
		_, _ = w.Write([]byte(`{
			"data": [` + campaignJSON + `],
			"pagination": {
				"token": 2,
				"prev_token": 1,
				"next_token": 3,
				"first_url": "https://mailtrap.io/api/email_campaigns?per_page=10&token=1",
				"prev_url": "https://mailtrap.io/api/email_campaigns?per_page=10&token=1",
				"current_url": "https://mailtrap.io/api/email_campaigns?per_page=10&token=2",
				"next_url": "https://mailtrap.io/api/email_campaigns?per_page=10&token=3"
			}
		}`))
	})

	opts := &mailtrap.EmailCampaignListOptions{Token: 2, PerPage: 10, Search: "spring"}
	list, _, err := client.EmailCampaigns.List(context.Background(), opts)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(list.Data))
	}
	c := list.Data[0]
	if c.ID != 4567 || c.Name != "Spring Sale" {
		t.Errorf("campaign = %+v", c)
	}
	if c.DomainID != 4321 {
		t.Errorf("DomainID = %d", c.DomainID)
	}
	if c.CurrentState != mailtrap.EmailCampaignStateDraft || c.DeliveryMode != mailtrap.EmailCampaignDeliveryModeRapid {
		t.Errorf("state = %q, mode = %q", c.CurrentState, c.DeliveryMode)
	}
	if c.RecipientTotalCount != nil {
		t.Errorf("RecipientTotalCount = %v, want nil", *c.RecipientTotalCount)
	}
	if len(c.ContactListIDs) != 2 || c.ContactListIDs[0] != 55 {
		t.Errorf("ContactListIDs = %v", c.ContactListIDs)
	}
	p := list.Pagination
	if p.Token != 2 || p.PrevToken == nil || *p.PrevToken != 1 || p.NextToken == nil || *p.NextToken != 3 {
		t.Errorf("pagination = %+v", p)
	}
}

func TestEmailCampaigns_All(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/email_campaigns", func(w http.ResponseWriter, r *http.Request) {
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
	for c, err := range client.EmailCampaigns.All(context.Background(), nil) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		ids = append(ids, c.ID)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Errorf("ids = %v", ids)
	}
}

func TestEmailCampaigns_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/email_campaigns/4567", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data": ` + campaignJSON + `}`))
	})

	c, _, err := client.EmailCampaigns.Get(context.Background(), 4567)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if c.ID != 4567 || c.DomainName != "acme.com" {
		t.Errorf("campaign = %+v", c)
	}
	if c.ReplyTo == nil || c.ReplyTo.LocalPart != "support" {
		t.Errorf("ReplyTo = %+v", c.ReplyTo)
	}
	if c.Template == nil {
		t.Fatalf("Template = nil")
	}
	if c.Template.ID != 789 || c.Template.BodyHTML == nil || *c.Template.BodyHTML != "<html></html>" {
		t.Errorf("Template = %+v", c.Template)
	}
	if c.Template.BodyText != nil {
		t.Errorf("BodyText = %v, want nil", *c.Template.BodyText)
	}
}

func TestEmailCampaigns_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/email_campaigns", func(w http.ResponseWriter, r *http.Request) {
		// The body must be flat — no {"email_campaign": ...} wrapper.
		wantJSONBody(t, r, `{
			"name": "Spring Sale",
			"domain_id": 4321,
			"from_display_name": "Acme Marketing",
			"from_local_part": "news",
			"reply_to": {"display_name": "Acme Support", "local_part": "support", "domain": "acme.com"},
			"template_attributes": {"subject": "Spring is here", "body_html": "<html></html>", "merge_tags": ["first_name"]},
			"delivery_mode": "gradual",
			"delivery_options": {"emails_per_hour": 1000},
			"contact_list_ids": [55, 56],
			"contact_segment_ids": [12]
		}`)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data": ` + campaignJSON + `}`))
	})

	req := &mailtrap.CreateEmailCampaignRequest{
		Name:            "Spring Sale",
		DomainID:        4321,
		FromDisplayName: "Acme Marketing",
		FromLocalPart:   "news",
		ReplyTo: &mailtrap.EmailCampaignReplyTo{
			DisplayName: "Acme Support",
			LocalPart:   "support",
			Domain:      "acme.com",
		},
		TemplateAttributes: &mailtrap.EmailCampaignTemplateAttributes{
			Subject:   "Spring is here",
			BodyHTML:  "<html></html>",
			MergeTags: []string{"first_name"},
		},
		DeliveryMode:      mailtrap.EmailCampaignDeliveryModeGradual,
		DeliveryOptions:   &mailtrap.EmailCampaignDeliveryOptions{EmailsPerHour: mailtrap.Ptr(1000)},
		ContactListIDs:    []int64{55, 56},
		ContactSegmentIDs: []int64{12},
	}
	c, resp, err := client.EmailCampaigns.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d", resp.StatusCode)
	}
	if c.ID != 4567 {
		t.Errorf("campaign = %+v", c)
	}
}

func TestEmailCampaigns_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/email_campaigns/4567", func(w http.ResponseWriter, r *http.Request) {
		// Partial update: only the set fields are sent.
		wantJSONBody(t, r, `{
			"name": "Spring Sale (updated)",
			"template_attributes": {"subject": "New subject"},
			"contact_list_ids": [55]
		}`)
		_, _ = w.Write([]byte(`{"data": ` + campaignJSON + `}`))
	})

	req := &mailtrap.UpdateEmailCampaignRequest{
		Name:               "Spring Sale (updated)",
		TemplateAttributes: &mailtrap.EmailCampaignTemplateAttributes{Subject: "New subject"},
		ContactListIDs:     &[]int64{55},
	}
	c, _, err := client.EmailCampaigns.Update(context.Background(), 4567, req)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if c.ID != 4567 {
		t.Errorf("campaign = %+v", c)
	}
}

func TestEmailCampaigns_UpdateClearAudience(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/email_campaigns/4567", func(w http.ResponseWriter, r *http.Request) {
		// An explicit empty slice must be serialized to clear all lists, while
		// the nil ContactSegmentIDs must be omitted to leave segments unchanged.
		wantJSONBody(t, r, `{"contact_list_ids": []}`)
		_, _ = w.Write([]byte(`{"data": ` + campaignJSON + `}`))
	})

	req := &mailtrap.UpdateEmailCampaignRequest{ContactListIDs: &[]int64{}}
	if _, _, err := client.EmailCampaigns.Update(context.Background(), 4567, req); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestEmailCampaigns_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/email_campaigns/4567", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.EmailCampaigns.Delete(context.Background(), 4567)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

func TestEmailCampaigns_Start(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/email_campaigns/4567/start", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data": {"id": 4567, "current_state": "started"}}`))
	})

	c, _, err := client.EmailCampaigns.Start(context.Background(), 4567)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if c.CurrentState != mailtrap.EmailCampaignStateStarted {
		t.Errorf("campaign = %+v", c)
	}
}

func TestEmailCampaigns_Schedule(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/email_campaigns/4567/schedule", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"datetime": "2026-06-01T09:00:00.000Z"}`)
		_, _ = w.Write([]byte(`{"data": {
			"id": 4567,
			"current_state": "scheduled",
			"current_state_metadata": {"scheduled_at": "2026-06-01T09:00:00.000Z"}
		}}`))
	})

	c, _, err := client.EmailCampaigns.Schedule(context.Background(), 4567, "2026-06-01T09:00:00.000Z")
	if err != nil {
		t.Fatalf("Schedule: %v", err)
	}
	if c.CurrentState != mailtrap.EmailCampaignStateScheduled {
		t.Errorf("campaign = %+v", c)
	}
	if c.CurrentStateMetadata.ScheduledAt != "2026-06-01T09:00:00.000Z" {
		t.Errorf("ScheduledAt = %q", c.CurrentStateMetadata.ScheduledAt)
	}
}

func TestEmailCampaigns_Cancel(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/email_campaigns/4567/cancel", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data": {"id": 4567, "current_state": "draft"}}`))
	})

	c, _, err := client.EmailCampaigns.Cancel(context.Background(), 4567)
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if c.CurrentState != mailtrap.EmailCampaignStateDraft {
		t.Errorf("campaign = %+v", c)
	}
}

func TestEmailCampaigns_Terminate(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/email_campaigns/4567/terminate", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data": {
			"id": 4567,
			"current_state": "terminating",
			"current_state_metadata": {"errors": [{"message": "Invalid recipient address", "rcpt_index": 0}]}
		}}`))
	})

	c, _, err := client.EmailCampaigns.Terminate(context.Background(), 4567)
	if err != nil {
		t.Fatalf("Terminate: %v", err)
	}
	if c.CurrentState != mailtrap.EmailCampaignStateTerminating {
		t.Errorf("campaign = %+v", c)
	}
	errs := c.CurrentStateMetadata.Errors
	if len(errs) != 1 || errs[0].Message != "Invalid recipient address" || errs[0].RcptIndex != 0 {
		t.Errorf("Errors = %+v", errs)
	}
}

func TestEmailCampaigns_Reset(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/email_campaigns/4567/reset", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data": {"id": 4567, "current_state": "draft"}}`))
	})

	c, _, err := client.EmailCampaigns.Reset(context.Background(), 4567)
	if err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if c.CurrentState != mailtrap.EmailCampaignStateDraft {
		t.Errorf("campaign = %+v", c)
	}
}

func TestEmailCampaigns_Stats(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/email_campaigns/4567/stats", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("start_date"); got != "2026-05-01" {
			t.Errorf("start_date = %q, want 2026-05-01", got)
		}
		if got := q.Get("end_date"); got != "2026-05-31" {
			t.Errorf("end_date = %q, want 2026-05-31", got)
		}
		_, _ = w.Write([]byte(`{"data": {
			"delivery_count": 1450,
			"open_count": 820,
			"click_count": 310,
			"bounce_count": 30,
			"unsubscription_count": 12,
			"sent_count": 1500,
			"spam_count": 5,
			"delivery_rate": 0.9667,
			"open_rate": 0.5655,
			"click_rate": 0.2138,
			"bounce_rate": 0.02,
			"spam_rate": 0.0033,
			"unsubscription_rate": 0.0083
		}}`))
	})

	opts := &mailtrap.EmailCampaignStatsOptions{StartDate: "2026-05-01", EndDate: "2026-05-31"}
	stats, _, err := client.EmailCampaigns.Stats(context.Background(), 4567, opts)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.DeliveryCount != 1450 || stats.SentCount != 1500 || stats.SpamCount != 5 {
		t.Errorf("stats = %+v", stats)
	}
	if stats.DeliveryRate != 0.9667 || stats.UnsubscriptionRate != 0.0083 {
		t.Errorf("rates = %+v", stats)
	}
}
