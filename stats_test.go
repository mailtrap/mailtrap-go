package mailtrap_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestStats_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/stats", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("start_date") != "2025-01-01" || q.Get("end_date") != "2025-12-31" {
			t.Errorf("date query = %v", q)
		}
		if got := q["domain_ids[]"]; !reflect.DeepEqual(got, []string{"3938", "3939"}) {
			t.Errorf("domain_ids[] = %v", got)
		}
		if got := q["sending_streams[]"]; !reflect.DeepEqual(got, []string{"transactional"}) {
			t.Errorf("sending_streams[] = %v", got)
		}
		_, _ = w.Write([]byte(`{"delivery_count":190,"delivery_rate":0.95,"bounce_count":10}`))
	})

	stats, _, err := client.Stats.Get(context.Background(), &mailtrap.StatsOptions{
		StartDate:        "2025-01-01",
		EndDate:          "2025-12-31",
		SendingDomainIDs: []int64{3938, 3939},
		SendingStreams:   []string{"transactional"},
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if stats.DeliveryCount != 190 || stats.DeliveryRate != 0.95 || stats.BounceCount != 10 {
		t.Errorf("stats = %+v", stats)
	}
}

func TestStats_ByDomain(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/stats/domains", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"domain_id":3938,"stats":{"delivery_count":100}}]`))
	})

	byDomain, _, err := client.Stats.ByDomain(context.Background(), &mailtrap.StatsOptions{
		StartDate: "2025-01-01",
		EndDate:   "2025-12-31",
	})
	if err != nil {
		t.Fatalf("ByDomain: %v", err)
	}
	if len(byDomain) != 1 || byDomain[0].DomainID != 3938 || byDomain[0].Stats.DeliveryCount != 100 {
		t.Errorf("byDomain = %+v", byDomain)
	}
}

func TestStats_ByCategory(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/stats/categories", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"category":"Welcome Email","stats":{"open_count":42}}]`))
	})

	byCategory, _, err := client.Stats.ByCategory(context.Background(), &mailtrap.StatsOptions{StartDate: "2025-01-01", EndDate: "2025-12-31"})
	if err != nil {
		t.Fatalf("ByCategory: %v", err)
	}
	if len(byCategory) != 1 || byCategory[0].Category != "Welcome Email" || byCategory[0].Stats.OpenCount != 42 {
		t.Errorf("byCategory = %+v", byCategory)
	}
}

func TestStats_ByEmailServiceProvider(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/stats/email_service_providers", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"email_service_provider":"Google","stats":{"click_count":7}}]`))
	})

	byESP, _, err := client.Stats.ByEmailServiceProvider(context.Background(), &mailtrap.StatsOptions{StartDate: "2025-01-01", EndDate: "2025-12-31"})
	if err != nil {
		t.Fatalf("ByEmailServiceProvider: %v", err)
	}
	if len(byESP) != 1 || byESP[0].EmailServiceProvider != "Google" || byESP[0].Stats.ClickCount != 7 {
		t.Errorf("byESP = %+v", byESP)
	}
}

func TestStats_ByDate(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/stats/date", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"date":"2025-01-01","stats":{"spam_count":1}}]`))
	})

	byDate, _, err := client.Stats.ByDate(context.Background(), &mailtrap.StatsOptions{StartDate: "2025-01-01", EndDate: "2025-12-31"})
	if err != nil {
		t.Fatalf("ByDate: %v", err)
	}
	if len(byDate) != 1 || byDate[0].Date != "2025-01-01" || byDate[0].Stats.SpamCount != 1 {
		t.Errorf("byDate = %+v", byDate)
	}
}
