package main

import (
	"context"
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

	usage, _, err := client.Billing.Usage(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if usage.Sending != nil && usage.Sending.Usage.SentMessagesCount != nil {
		sent := usage.Sending.Usage.SentMessagesCount
		fmt.Printf("email sending: %d/%d messages this cycle\n", sent.Current, sent.Limit)
	}
	if usage.Testing != nil && usage.Testing.Usage.SentMessagesCount != nil {
		sent := usage.Testing.Usage.SentMessagesCount
		fmt.Printf("sandbox: %d/%d messages this cycle\n", sent.Current, sent.Limit)
	}
}
