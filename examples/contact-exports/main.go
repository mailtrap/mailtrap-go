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

	// Filters are optional; pass nil to export every contact.
	export, _, err := client.ContactExports.Create(ctx, []*mailtrap.ContactExportFilter{
		{Name: mailtrap.ContactExportFilterListID, Operator: mailtrap.ContactExportOperatorEqual, Value: []int64{101, 102}},
		{Name: mailtrap.ContactExportFilterSubscriptionStatus, Operator: mailtrap.ContactExportOperatorEqual, Value: mailtrap.ContactStatusSubscribed},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("started export %d (%s)\n", export.ID, export.Status)

	// Poll until the export finishes, giving up after a minute.
	pollCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	for export.Status != mailtrap.ContactExportFinished {
		time.Sleep(2 * time.Second)
		export, _, err = client.ContactExports.Get(pollCtx, export.ID)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("export ready: %s\n", *export.URL)
}
