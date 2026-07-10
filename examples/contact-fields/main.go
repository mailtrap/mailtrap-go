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

	field, _, err := client.ContactFields.Create(ctx, &mailtrap.CreateContactFieldRequest{
		Name:     "Age",
		DataType: mailtrap.ContactFieldTypeInteger,
		MergeTag: "age",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created field %d (%s, %s)\n", field.ID, field.Name, field.DataType)

	fields, _, err := client.ContactFields.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("account has %d field(s)\n", len(fields))

	// The data type is immutable; only name and merge tag can change.
	if _, _, err = client.ContactFields.Update(ctx, field.ID, &mailtrap.UpdateContactFieldRequest{
		Name:     "Years",
		MergeTag: "years",
	}); err != nil {
		log.Fatal(err)
	}

	if _, err = client.ContactFields.Delete(ctx, field.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted field")
}
