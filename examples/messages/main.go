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
	sandboxID, _ := strconv.ParseInt(os.Getenv("MAILTRAP_SANDBOX_ID"), 10, 64)

	client, err := mailtrap.NewClient(apiToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Iterate every message with the auto-paging iterator...
	for message, err := range client.SandboxMessages.All(ctx, sandboxID, nil) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("message %d: %s\n", message.ID, message.Subject)
	}

	// ...or fetch a single page yourself, optionally filtered.
	messages, _, err := client.SandboxMessages.List(ctx, sandboxID, &mailtrap.MessageListOptions{Search: "welcome"})
	if err != nil {
		log.Fatal(err)
	}
	if len(messages) == 0 {
		return
	}
	messageID := messages[0].ID

	message, _, err := client.SandboxMessages.Get(ctx, sandboxID, messageID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("subject: %s\n", message.Subject)

	_, _, err = client.SandboxMessages.Update(ctx, sandboxID, messageID, true)
	if err != nil {
		log.Fatal(err)
	}

	report, _, err := client.SandboxMessages.SpamReport(ctx, sandboxID, messageID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("spam score: %.1f\n", report.Score)

	text, _, err := client.SandboxMessages.Text(ctx, sandboxID, messageID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("text body: %d bytes\n", len(text))

	_, err = client.SandboxMessages.Forward(ctx, sandboxID, messageID, "qa@example.com")
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = client.SandboxMessages.Delete(ctx, sandboxID, messageID)
	if err != nil {
		log.Fatal(err)
	}
}
