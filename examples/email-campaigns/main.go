package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mailtrap/mailtrap-go"
)

func main() {
	contactListID, _ := strconv.ParseInt(os.Getenv("MAILTRAP_CONTACT_LIST_ID"), 10, 64)
	domainID, _ := strconv.ParseInt(os.Getenv("MAILTRAP_DOMAIN_ID"), 10, 64)

	client, err := mailtrap.NewClient(os.Getenv("MAILTRAP_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create a draft campaign on a verified sending domain.
	campaign, _, err := client.EmailCampaigns.Create(ctx, &mailtrap.CreateEmailCampaignRequest{
		Name:            "Spring Sale",
		DomainID:        domainID,
		FromDisplayName: "Acme Marketing",
		FromLocalPart:   "news",
		TemplateAttributes: &mailtrap.EmailCampaignTemplateAttributes{
			Subject: "Spring is here — 30% off",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created campaign %d (%s), state %s\n", campaign.ID, campaign.Name, campaign.CurrentState)

	// Add the design and the audience, and throttle delivery.
	campaign, _, err = client.EmailCampaigns.Update(ctx, campaign.ID, &mailtrap.UpdateEmailCampaignRequest{
		TemplateAttributes: &mailtrap.EmailCampaignTemplateAttributes{
			BodyHTML:  `<html><body><h1>Hi {{first_name}}!</h1><p><a href="__unsubscribe_url__">Unsubscribe</a></p></body></html>`,
			MergeTags: []string{"first_name"},
		},
		ContactListIDs:  &[]int64{contactListID},
		DeliveryMode:    mailtrap.EmailCampaignDeliveryModeGradual,
		DeliveryOptions: &mailtrap.EmailCampaignDeliveryOptions{EmailsPerHour: mailtrap.Ptr(1000)},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("updated design + audience; subject %q\n", campaign.Template.Subject)

	// Schedule the campaign for tomorrow (the time must be in the future and
	// at most one month ahead), then change plans and cancel it back to draft.
	datetime := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02T15:04:05.000Z")
	campaign, _, err = client.EmailCampaigns.Schedule(ctx, campaign.ID, datetime)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("scheduled for %s\n", campaign.CurrentStateMetadata.ScheduledAt)

	if campaign, _, err = client.EmailCampaigns.Cancel(ctx, campaign.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("cancelled, back to %s\n", campaign.CurrentState)

	// Or start sending immediately.
	if campaign, _, err = client.EmailCampaigns.Start(ctx, campaign.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("started, state %s\n", campaign.CurrentState)

	// Fetch a single campaign by ID.
	campaign, _, err = client.EmailCampaigns.Get(ctx, campaign.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("campaign %d targets %v\n", campaign.ID, campaign.ContactListIDs)

	// List one page of campaigns, filtered by name.
	page, _, err := client.EmailCampaigns.List(ctx, &mailtrap.EmailCampaignListOptions{PerPage: 10, Search: "spring"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("page has %d campaign(s); next token: %v\n", len(page.Data), page.Pagination.NextToken)

	// Or iterate every campaign across pages.
	for c, err := range client.EmailCampaigns.All(ctx, nil) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("- %d %s (%s)\n", c.ID, c.Name, c.CurrentState)
	}

	// Aggregated performance metrics, optionally narrowed to a date window.
	stats, _, err := client.EmailCampaigns.Stats(ctx, campaign.ID, &mailtrap.EmailCampaignStatsOptions{
		StartDate: "2026-05-01",
		EndDate:   "2026-05-31",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sent %d, delivered %d (%.2f%%), opened %d\n",
		stats.SentCount, stats.DeliveryCount, stats.DeliveryRate*100, stats.OpenCount)

	// Delete the campaign (204 No Content).
	if _, err := client.EmailCampaigns.Delete(ctx, campaign.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("campaign deleted")
}
