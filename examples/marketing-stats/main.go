package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mailtrap/mailtrap-go"
)

func main() {
	client, err := mailtrap.NewClient(os.Getenv("MAILTRAP_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	now := time.Now().UTC()
	opts := &mailtrap.MarketingStatsOptions{
		StartDate: now.AddDate(0, 0, -30).Format(time.DateOnly),
		EndDate:   now.Format(time.DateOnly),
	}

	// Campaigns and automations together.
	stats, _, err := client.MarketingStats.Get(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	printStats("all", stats)

	// Campaigns only.
	opts.Sources = []string{mailtrap.MarketingStatsSourceCampaigns}
	stats, _, err = client.MarketingStats.Get(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	printStats("campaigns", stats)
}

func printStats(label string, s *mailtrap.MarketingStats) {
	fmt.Printf("%s: delivered %d (%.0f%%), opened %d (%.0f%%)\n", label, s.DeliveryCount, s.DeliveryRate*100, s.OpenCount, s.OpenRate*100)
}
