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

	webhook, _, err := client.Webhooks.Create(ctx, &mailtrap.CreateWebhookRequest{
		URL:         "https://example.com/mailtrap/webhooks",
		WebhookType: mailtrap.WebhookTypeEmailSending,
		EventTypes:  []string{mailtrap.WebhookEventDelivery, mailtrap.WebhookEventBounce, mailtrap.WebhookEventOpen},
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
