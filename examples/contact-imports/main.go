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

	imp, _, err := client.ContactImports.Create(ctx, []*mailtrap.ImportContact{
		{Email: "customer1@example.com", Fields: map[string]any{"first_name": "John"}, ListIDsIncluded: []int64{1, 2}},
		{Email: "customer2@example.com", Fields: map[string]any{"first_name": "Joe"}, ListIDsIncluded: []int64{1}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("started import %d (%s)\n", imp.ID, imp.Status)

	// Poll until the job reaches a terminal status, giving up after a minute.
	pollCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	for imp.Status != mailtrap.ContactImportFinished && imp.Status != mailtrap.ContactImportFailed {
		time.Sleep(2 * time.Second)
		imp, _, err = client.ContactImports.Get(pollCtx, imp.ID)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("import %s: %d created, %d updated\n", imp.Status, imp.CreatedContactsCount, imp.UpdatedContactsCount)
}
