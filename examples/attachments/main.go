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

	// Get the first message in the sandbox.
	messages, _, err := client.SandboxMessages.List(ctx, sandboxID, nil)
	if err != nil {
		log.Fatal(err)
	}
	if len(messages) == 0 {
		return
	}
	messageID := messages[0].ID

	attachments, _, err := client.SandboxAttachments.List(ctx, sandboxID, messageID, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("message has %d attachment(s)\n", len(attachments))

	for _, a := range attachments {
		full, _, err := client.SandboxAttachments.Get(ctx, sandboxID, messageID, a.ID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s (%s)\n", full.Filename, full.ContentType)
	}
}
