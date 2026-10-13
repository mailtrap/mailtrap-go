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
	now := time.Now()
	opts := &mailtrap.MarketingStatsOptions{
		StartDate: now.AddDate(0, 0, -30).Format(time.DateOnly),
		EndDate:   now.Format(time.DateOnly),
	}

	// Campaigns and automations together.
	stats, _, err := client.MarketingStats.Get(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("all: delivered %d (%.0f%%), opened %d (%.0f%%)\n", stats.DeliveryCount, stats.DeliveryRate*100, stats.OpenCount, stats.OpenRate*100)

	// Campaigns only.
	opts.Sources = []string{mailtrap.MarketingStatsSourceCampaigns}
	stats, _, err = client.MarketingStats.Get(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("campaigns: delivered %d (%.0f%%), opened %d (%.0f%%)\n", stats.DeliveryCount, stats.DeliveryRate*100, stats.OpenCount, stats.OpenRate*100)
}
