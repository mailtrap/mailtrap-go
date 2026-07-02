package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mailtrap/mailtrap-go"
)

const apiToken = "your-api-token"

func main() {
	client, err := mailtrap.NewClient(apiToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	opts := &mailtrap.StatsOptions{
		StartDate: "2025-01-01",
		EndDate:   "2025-12-31",
	}

	stats, _, err := client.Stats.Get(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("delivered %d (%.0f%%), bounced %d\n", stats.DeliveryCount, stats.DeliveryRate*100, stats.BounceCount)

	byDomain, _, err := client.Stats.ByDomain(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range byDomain {
		fmt.Printf("domain %d: %d delivered\n", d.DomainID, d.Stats.DeliveryCount)
	}

	byDate, _, err := client.Stats.ByDate(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range byDate {
		fmt.Printf("%s: %d opened\n", d.Date, d.Stats.OpenCount)
	}
}
