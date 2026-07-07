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

	if _, _, err = client.Suppressions.Delete(ctx, suppression.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("removed suppression")
}
