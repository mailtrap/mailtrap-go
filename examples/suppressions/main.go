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

	suppression, _, err := client.Suppressions.Create(ctx, &mailtrap.CreateSuppressionRequest{
		Email:         "bounced@example.com",
		DomainID:      12345,
		SendingStream: "transactional",
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

	if _, _, err = client.Suppressions.Delete(ctx, suppression.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("removed suppression")
}
