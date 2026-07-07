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
	projectID, _ := strconv.ParseInt(os.Getenv("MAILTRAP_PROJECT_ID"), 10, 64)

	client, err := mailtrap.NewClient(apiToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	accesses, _, err := client.AccountAccesses.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, access := range accesses {
		fmt.Printf("access %d: %s\n", access.ID, access.SpecifierType)
	}

	// Filter to accesses on specific projects, sandboxes, or domains.
	_, _, err = client.AccountAccesses.List(ctx, &mailtrap.AccountAccessListOptions{
		ProjectIDs: []int64{projectID},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Remove an access by ID.
	if len(accesses) > 0 {
		if _, err := client.AccountAccesses.Delete(ctx, accesses[0].ID); err != nil {
			log.Fatal(err)
		}
	}
}
