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

	contact, _, err := client.Contacts.Create(ctx, &mailtrap.CreateContactRequest{
		Email: "john.smith@example.com",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Record a custom event for the contact (by UUID or email).
	event, _, err := client.ContactEvents.Create(ctx, contact.ID, &mailtrap.CreateContactEventRequest{
		Name:   "UserLogin",
		Params: map[string]any{"user_id": 101, "is_active": true},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("recorded event %q for %s\n", event.Name, event.ContactEmail)
}
