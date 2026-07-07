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
	organizationID, err := strconv.ParseInt(os.Getenv("MAILTRAP_ORGANIZATION_ID"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	client, err := mailtrap.NewClient(os.Getenv("MAILTRAP_API_TOKEN"),
		mailtrap.WithOrganizationID(organizationID),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	subAccount, _, err := client.SubAccounts.Create(ctx, "Development Team Account")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created sub-account %d (%s)\n", subAccount.ID, subAccount.Name)

	subAccounts, _, err := client.SubAccounts.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("organization has %d sub-account(s)\n", len(subAccounts))
}
