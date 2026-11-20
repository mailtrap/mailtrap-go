package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/mailtrap/mailtrap-go"
)

func main() {
	apiToken := os.Getenv("MAILTRAP_API_TOKEN")
	domainID, _ := strconv.ParseInt(os.Getenv("MAILTRAP_DOMAIN_ID"), 10, 64)

	client, err := mailtrap.NewClient(apiToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	suppression, _, err := client.Suppressions.Create(ctx, &mailtrap.CreateSuppressionRequest{
		Email:         "bounced@example.com",
		DomainID:      domainID,
		SendingStream: mailtrap.SendingStreamTransactional,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("suppressed %s (%s)\n", suppression.Email, suppression.ID)

	suppressions, _, err := client.Suppressions.List(ctx, &mailtrap.SuppressionListOptions{
		Email: "bounced@example.com",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("found %d suppression(s)\n", len(suppressions))

	// Also list recipients' unsubscribes from single subscription topics. They
	// can't be deleted, so the cleanup below removes only the suppression
	// created above.
	suppressions, _, err = client.Suppressions.List(ctx, &mailtrap.SuppressionListOptions{
		IncludeTopicUnsubscriptions: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range suppressions {
		if s.Type == mailtrap.SuppressionTypeTopicUnsubscription && s.SubscriptionTopicID != nil {
			fmt.Printf("%s unsubscribed from topic %d on %s\n", s.Email, *s.SubscriptionTopicID, s.DomainName)
		}
	}

	if _, _, err = client.Suppressions.Delete(ctx, suppression.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("removed suppression")
}
