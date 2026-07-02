package main

import (
	"context"
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

	webhook, _, err := client.Webhooks.Create(ctx, &mailtrap.CreateWebhookRequest{
		URL:         "https://example.com/mailtrap/webhooks",
		WebhookType: "email_sending",
		EventTypes:  []string{"delivery", "bounce", "open"},
	})
	if err != nil {
		log.Fatal(err)
	}
	// SigningSecret is only returned here — store it to verify payload signatures.
	fmt.Printf("created webhook %d, signing secret: %s\n", webhook.ID, webhook.SigningSecret)

	webhooks, _, err := client.Webhooks.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("account has %d webhook(s)\n", len(webhooks))

	_, _, err = client.Webhooks.Update(ctx, webhook.ID, &mailtrap.UpdateWebhookRequest{
		Active: mailtrap.Ptr(false),
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, _, err = client.Webhooks.Delete(ctx, webhook.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted webhook")
}
