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

	ctx := context.Background()
	opts := &mailtrap.EmailLogsListOptions{
		SentAfter:  "2025-01-01T00:00:00Z",
		SentBefore: "2025-01-31T23:59:59Z",
		Filters: map[string]mailtrap.LogFilter{
			"status": {Operator: "equal", Values: []string{mailtrap.EmailLogStatusDelivered}},
			"to":     {Operator: "ci_contain", Values: []string{"@example.com"}},
		},
	}

	list, _, err := client.EmailLogs.List(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d message(s) match; showing %d\n", list.TotalCount, len(list.Messages))

	// Iterate every match, following the cursor across pages.
	for msg, err := range client.EmailLogs.All(ctx, opts) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s  %s -> %s  [%s]\n", msg.SentAt, msg.From, msg.To, msg.Status)
	}

	if len(list.Messages) == 0 {
		return
	}

	msg, _, err := client.EmailLogs.Get(ctx, list.Messages[0].MessageID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("message %s has %d event(s)\n", msg.MessageID, len(msg.Events))
	for _, e := range msg.Events {
		fmt.Printf("  %s at %s\n", e.EventType, e.CreatedAt)
	}
}
