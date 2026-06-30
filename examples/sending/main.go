package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/mailtrap/mailtrap-go"
)

const apiToken = "your-api-token"

func main() {
	client, err := mailtrap.NewClient(apiToken)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Transactional send.
	resp, _, err := client.Send(ctx, &mailtrap.SendRequest{
		From:            mailtrap.Address{Email: "sender@example.com", Name: "Example"},
		To:              []mailtrap.Address{{Email: "recipient@example.com"}},
		Cc:              []mailtrap.Address{{Email: "cc@example.com"}},
		ReplyTo:         &mailtrap.Address{Email: "support@example.com"},
		Subject:         "Your order confirmation",
		Text:            "Thanks for your order!",
		HTML:            "<h1>Thanks for your order!</h1>",
		Category:        "order-confirmation",
		CustomVariables: map[string]any{"order_id": "ORD-789"},
		Headers:         map[string]string{"X-Message-Source": "api.example.com"},
		Attachments: []mailtrap.Attachment{{
			Filename: "receipt.txt",
			Type:     "text/plain",
			Content:  base64.StdEncoding.EncodeToString([]byte("Thank you for your order!")),
		}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("sent: %v %v\n", resp.Success, resp.MessageIDs)

	// Template send: subject and body come from the template.
	tmpl, _, err := client.Send(ctx, &mailtrap.SendRequest{
		From:              mailtrap.Address{Email: "sender@example.com"},
		To:                []mailtrap.Address{{Email: "recipient@example.com"}},
		TemplateUUID:      "b81aabcd-1a1e-41cf-91b6-eca0254b3d96",
		TemplateVariables: map[string]any{"user_name": "Jane", "order_number": "12345"},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("template sent: %v\n", tmpl.Success)

	// Bulk stream for high-volume sending; selected with WithBulk at construction.
	bulkClient, err := mailtrap.NewClient(apiToken, mailtrap.WithBulk(true))
	if err != nil {
		log.Fatal(err)
	}
	bulk, _, err := bulkClient.Send(ctx, &mailtrap.SendRequest{
		From:     mailtrap.Address{Email: "marketing@example.com"},
		To:       []mailtrap.Address{{Email: "subscriber@example.com"}},
		Subject:  "Monthly newsletter",
		HTML:     "<h1>Our latest updates</h1>",
		Category: "newsletter",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bulk sent: %v\n", bulk.Success)

	// Batch send: Base holds shared fields; per-request fields override it.
	batch, _, err := client.SendBatch(ctx, &mailtrap.BatchSendRequest{
		Base: &mailtrap.SendRequest{
			From:    mailtrap.Address{Email: "sender@example.com"},
			Subject: "Hello {{name}}",
			Text:    "Hello {{name}}, we have news for you.",
		},
		Requests: []mailtrap.SendRequest{
			{To: []mailtrap.Address{{Email: "user1@example.com"}}, CustomVariables: map[string]any{"name": "Alice"}},
			{To: []mailtrap.Address{{Email: "user2@example.com"}}, CustomVariables: map[string]any{"name": "Bob"}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("batch sent: %v (%d responses)\n", batch.Success, len(batch.Responses))
}
