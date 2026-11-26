package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/mailtrap/mailtrap-go"
)

func main() {
	client, err := mailtrap.NewClient(os.Getenv("MAILTRAP_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// List one page of topics. Topics are managed in the Mailtrap app; the API
	// reads them so their IDs can be set on campaigns and contacts.
	page, _, err := client.SubscriptionTopics.List(ctx, &mailtrap.SubscriptionTopicListOptions{PerPage: 10})
	var forbidden *mailtrap.ForbiddenError
	if errors.As(err, &forbidden) {
		// Either the account has no subscription topics or the token can't read them.
		fmt.Println("subscription topics are not available for this account or token:", err)
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("page has %d topic(s)\n", len(page.Data))
	if next := page.Pagination.NextToken; next != nil {
		fmt.Printf("next page token: %d\n", *next)
	}

	// Or iterate every topic across pages. Deleted topics are left out.
	for t, err := range client.SubscriptionTopics.All(ctx, nil) {
		if err != nil {
			log.Fatal(err)
		}
		scope := "every domain"
		if t.DomainID != nil {
			scope = fmt.Sprintf("domain %d", *t.DomainID)
		}
		fmt.Printf("- %d %s (%s)\n", t.ID, t.Name, scope)
	}

	if len(page.Data) == 0 {
		return
	}

	// Get a topic by ID. Get also returns a deleted topic, with DeletedAt set,
	// since a campaign keeps the topic it was given.
	topic, _, err := client.SubscriptionTopics.Get(ctx, page.Data[0].ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("topic %d %q, deleted: %t\n", topic.ID, topic.Name, topic.DeletedAt != nil)
}
