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

	list, _, err := client.ContactLists.Create(ctx, "Customers")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created list %d (%s)\n", list.ID, list.Name)

	lists, _, err := client.ContactLists.List(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("account has %d list(s)\n", len(lists))

	// Filter lists by name.
	filtered, _, err := client.ContactLists.List(ctx, &mailtrap.ContactListListOptions{Search: "cust"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d list(s) match %q\n", len(filtered), "cust")

	if _, _, err = client.ContactLists.Update(ctx, list.ID, "Former Customers"); err != nil {
		log.Fatal(err)
	}

	if _, err = client.ContactLists.Delete(ctx, list.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted list")
}
