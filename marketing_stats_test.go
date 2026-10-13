package mailtrap_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestMarketingStats_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/marketing/stats", func(w http.ResponseWriter, r *http.Request) {
		want := map[string][]string{
			"start_date":                {"2026-05-01"},
			"end_date":                  {"2026-05-31"},
			"sending_domain_ids[]":      {"3938", "3939"},
			"sources[]":                 {"campaigns"},
			"email_service_providers[]": {"Google", "Yahoo"},
		}
		if got := map[string][]string(r.URL.Query()); !reflect.DeepEqual(got, want) {
			t.Errorf("query = %v, want %v", got, want)
		}
		_, _ = w.Write([]byte(`{"data": {
			"delivery_count": 1450,
			"delivery_rate": 0.9667,
			"bounce_count": 30,
			"bounce_rate": 0.02,
			"open_count": 820,
			"open_rate": 0.5655,
			"click_count": 310,
			"click_rate": 0.2138,
			"spam_count": 5,
			"spam_rate": 0.0033,
			"unsubscription_count": 12,
			"unsubscription_rate": 0.0083,
			"sent_count": 1500
		}}`))
	})

	stats, _, err := client.MarketingStats.Get(context.Background(), &mailtrap.MarketingStatsOptions{
		StartDate:             "2026-05-01",
		EndDate:               "2026-05-31",
		SendingDomainIDs:      []int64{3938, 3939},
		Sources:               []string{mailtrap.MarketingStatsSourceCampaigns},
		EmailServiceProviders: []string{"Google", "Yahoo"},
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	want := &mailtrap.MarketingStats{
		DeliveryCount: 1450, DeliveryRate: 0.9667,
		BounceCount: 30, BounceRate: 0.02,
		OpenCount: 820, OpenRate: 0.5655,
		ClickCount: 310, ClickRate: 0.2138,
		SpamCount: 5, SpamRate: 0.0033,
		UnsubscriptionCount: 12, UnsubscriptionRate: 0.0083,
		SentCount: 1500,
	}
	if !reflect.DeepEqual(stats, want) {
		t.Errorf("stats = %+v, want %+v", stats, want)
	}
}

func TestMarketingStats_Get_DatesOnly(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/marketing/stats", func(w http.ResponseWriter, r *http.Request) {
		// Omitted filters send no keys.
		want := map[string][]string{"start_date": {"2026-05-01"}, "end_date": {"2026-05-31"}}
		if got := map[string][]string(r.URL.Query()); !reflect.DeepEqual(got, want) {
			t.Errorf("query = %v, want %v", got, want)
		}
		_, _ = w.Write([]byte(`{"data": {}}`))
	})

	stats, _, err := client.MarketingStats.Get(context.Background(), &mailtrap.MarketingStatsOptions{
		StartDate: "2026-05-01",
		EndDate:   "2026-05-31",
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !reflect.DeepEqual(stats, &mailtrap.MarketingStats{}) {
		t.Errorf("stats = %+v, want zero value", stats)
	}
}
