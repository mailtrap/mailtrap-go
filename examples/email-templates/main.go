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

	template, _, err := client.EmailTemplates.Create(ctx, &mailtrap.EmailTemplateRequest{
		Name:     "Welcome",
		Subject:  "Welcome aboard",
		Category: "Onboarding",
		BodyHTML: "<h1>Welcome!</h1>",
		BodyText: "Welcome!",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created template %d (%s)\n", template.ID, template.UUID)

	templates, _, err := client.EmailTemplates.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("account has %d template(s)\n", len(templates))

	_, _, err = client.EmailTemplates.Update(ctx, template.ID, &mailtrap.EmailTemplateRequest{
		Subject: "Welcome to Mailtrap",
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err = client.EmailTemplates.Delete(ctx, template.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted template")
}
