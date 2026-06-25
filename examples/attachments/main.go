package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mailtrap/mailtrap-go"
)

const (
	apiToken  = "your-api-token"
	sandboxID = 3000001
	messageID = 4000001
)

func main() {
	client, err := mailtrap.NewClient(apiToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

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
