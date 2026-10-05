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

	template, _, err := client.Templates.Create(ctx, &mailtrap.CreateTemplateRequest{
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

	// List one page of templates.
	page, _, err := client.Templates.List(ctx, &mailtrap.TemplateListOptions{PerPage: 10})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("page has %d template(s); next token: %v\n", len(page.Data), page.Pagination.NextToken)

	// Or iterate every template across pages.
	for t, err := range client.Templates.All(ctx, nil) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("- %d %s (%s)\n", t.ID, t.Name, t.Category)
	}

	template, _, err = client.Templates.Update(ctx, template.ID, &mailtrap.UpdateTemplateRequest{
		Subject: "Welcome to Mailtrap",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("updated subject: %s\n", template.Subject)

	if _, err = client.Templates.Delete(ctx, template.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted template")
}
