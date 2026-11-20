package mailtrap_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

const subscriptionTopicJSON = `{
	"id": 12,
	"name": "Product updates",
	"description": "News about features we ship",
	"domain_id": 4321,
	"deleted_at": null,
	"created_at": "2026-05-01T10:15:00.000Z",
	"updated_at": "2026-05-02T09:00:00.000Z"
}`

func TestSubscriptionTopics_List(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/subscription_topics", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RawQuery; got != "per_page=10&token=2" {
			t.Errorf("query = %q, want per_page=10&token=2", got)
		}
		_, _ = w.Write([]byte(`{
			"data": [` + subscriptionTopicJSON + `, {
				"id": 13,
				"name": "Events",
				"description": null,
				"domain_id": null,
				"deleted_at": null,
				"created_at": "2026-05-03T10:15:00.000Z",
				"updated_at": "2026-05-03T10:15:00.000Z"
			}],
			"pagination": {
				"token": 2,
				"prev_token": 1,
				"next_token": 3,
				"first_url": "https://mailtrap.io/api/subscription_topics?per_page=10&token=1",
				"prev_url": "https://mailtrap.io/api/subscription_topics?per_page=10&token=1",
				"current_url": "https://mailtrap.io/api/subscription_topics?per_page=10&token=2",
				"next_url": "https://mailtrap.io/api/subscription_topics?per_page=10&token=3"
			}
		}`))
	})

	opts := &mailtrap.SubscriptionTopicListOptions{Token: 2, PerPage: 10}
	list, _, err := client.SubscriptionTopics.List(context.Background(), opts)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(list.Data))
	}
	topic := list.Data[0]
	if topic.ID != 12 || topic.Name != "Product updates" || topic.CreatedAt != "2026-05-01T10:15:00.000Z" {
		t.Errorf("topic = %+v", topic)
	}
	if topic.Description == nil || *topic.Description != "News about features we ship" {
		t.Errorf("Description = %v", topic.Description)
	}
	if topic.DomainID == nil || *topic.DomainID != 4321 {
		t.Errorf("DomainID = %v, want 4321", topic.DomainID)
	}
	if topic.DeletedAt != nil {
		t.Errorf("DeletedAt = %v, want nil", *topic.DeletedAt)
	}
	// A topic for every domain has no domain_id, and description is optional.
	if all := list.Data[1]; all.DomainID != nil || all.Description != nil {
		t.Errorf("all-domains topic = %+v", all)
	}
	p := list.Pagination
	if p.Token != 2 || p.PrevToken == nil || *p.PrevToken != 1 || p.NextToken == nil || *p.NextToken != 3 {
		t.Errorf("pagination = %+v", p)
	}
}

func TestSubscriptionTopics_ListDefaults(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/subscription_topics", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RawQuery; got != "" {
			t.Errorf("query = %q, want none", got)
		}
		_, _ = w.Write([]byte(`{"data": [], "pagination": {"token": 1, "prev_token": null, "next_token": null}}`))
	})

	list, _, err := client.SubscriptionTopics.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list.Data) != 0 || list.Pagination.NextToken != nil {
		t.Errorf("list = %+v", list)
	}
}

func TestSubscriptionTopics_All(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/subscription_topics", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("per_page"); got != "2" {
			t.Errorf("per_page = %q, want 2 on every page", got)
		}
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
	opts := &mailtrap.SubscriptionTopicListOptions{PerPage: 2}
	for topic, err := range client.SubscriptionTopics.All(context.Background(), opts) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		ids = append(ids, topic.ID)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Errorf("ids = %v", ids)
	}
	if opts.Token != 0 {
		t.Errorf("All changed the caller's options: Token = %d", opts.Token)
	}
}

func TestSubscriptionTopics_AllStopsAtError(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/subscription_topics", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") == "" {
			_, _ = w.Write([]byte(`{
				"data": [{"id": 1, "name": "First"}],
				"pagination": {"token": 1, "prev_token": null, "next_token": 2}
			}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	})

	var ids []int64
	var errs []error
	for topic, err := range client.SubscriptionTopics.All(context.Background(), nil) {
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ids = append(ids, topic.ID)
	}
	if len(ids) != 1 || ids[0] != 1 {
		t.Errorf("ids = %v, want [1]", ids)
	}
	var apiErr *mailtrap.Error
	if len(errs) != 1 || !errors.As(errs[0], &apiErr) || apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("errs = %v, want one 500 *Error", errs)
	}
}

func TestSubscriptionTopics_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/subscription_topics/12", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data": ` + subscriptionTopicJSON + `}`))
	})

	topic, _, err := client.SubscriptionTopics.Get(context.Background(), 12)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if topic.ID != 12 || topic.Name != "Product updates" || topic.UpdatedAt != "2026-05-02T09:00:00.000Z" {
		t.Errorf("topic = %+v", topic)
	}
	if topic.DeletedAt != nil {
		t.Errorf("DeletedAt = %v, want nil", *topic.DeletedAt)
	}
}

func TestSubscriptionTopics_GetDeleted(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/subscription_topics/14", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data": {
			"id": 14,
			"name": "Old newsletter",
			"description": null,
			"domain_id": null,
			"deleted_at": "2026-06-01T08:00:00.000Z",
			"created_at": "2026-01-01T08:00:00.000Z",
			"updated_at": "2026-06-01T08:00:00.000Z"
		}}`))
	})

	topic, _, err := client.SubscriptionTopics.Get(context.Background(), 14)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if topic.DeletedAt == nil || *topic.DeletedAt != "2026-06-01T08:00:00.000Z" {
		t.Errorf("DeletedAt = %v", topic.DeletedAt)
	}
}

func TestSubscriptionTopics_Forbidden(t *testing.T) {
	mux, client := setup(t)
	// An account without subscription topics gets a 403 with an empty body.
	mux.HandleFunc("GET /api/subscription_topics", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	_, _, err := client.SubscriptionTopics.List(context.Background(), nil)
	var forbidden *mailtrap.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("err = %v, want *ForbiddenError", err)
	}
	if forbidden.Err.StatusCode != http.StatusForbidden || len(forbidden.Err.Messages) != 0 {
		t.Errorf("err = %+v", forbidden.Err)
	}
}
