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

	client, err := mailtrap.NewClient(apiToken,
		mailtrap.WithSandbox(true),
		mailtrap.WithSandboxID(sandboxID),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	resp, _, err := client.Send(ctx, &mailtrap.SendRequest{
		From:    mailtrap.Address{Email: "sender@example.com", Name: "Example"},
		To:      []mailtrap.Address{{Email: "recipient@example.com"}},
		Subject: "Hello from mailtrap-go",
		Text:    "This message was captured by the sandbox.",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sent: %v %v\n", resp.Success, resp.MessageIDs)

	batch, _, err := client.SendBatch(ctx, &mailtrap.BatchSendRequest{
		Base: &mailtrap.SendRequest{
			From:    mailtrap.Address{Email: "sender@example.com"},
			Subject: "Batch hello",
			Text:    "Sent via batch.",
		},
		Requests: []mailtrap.SendRequest{
			{To: []mailtrap.Address{{Email: "a@example.com"}}},
			{To: []mailtrap.Address{{Email: "b@example.com"}}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("batch sent: %v (%d responses)\n", batch.Success, len(batch.Responses))
}
