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
		Email:   "john.smith@example.com",
		Fields:  map[string]any{"first_name": "John", "last_name": "Smith"},
		ListIDs: []int64{1, 2, 3},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created contact %s (%s)\n", contact.ID, contact.Status)

	// Get by UUID or email (the email is URL-encoded for you).
	got, _, err := client.Contacts.Get(ctx, contact.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("contact email: %s\n", got.Email)

	// Update is an upsert; it reports whether the contact was created or updated.
	upsert, _, err := client.Contacts.Update(ctx, contact.ID, &mailtrap.UpdateContactRequest{
		Email:        "john.smith@example.com",
		Fields:       map[string]any{"first_name": "Johnny"},
		Unsubscribed: mailtrap.Ptr(true),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("contact %s\n", upsert.Action)

	if _, err := client.Contacts.Delete(ctx, contact.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted contact")
}
