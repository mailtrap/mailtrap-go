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

	domain, _, err := client.SendingDomains.Create(ctx, "example.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created domain %d (%s), status %s\n", domain.ID, domain.DomainName, domain.ComplianceStatus)

	for _, rec := range domain.DNSRecords {
		fmt.Printf("  %s %s %q -> %q (%s)\n", rec.Key, rec.Type, rec.Name, rec.Value, rec.Status)
	}

	domains, _, err := client.SendingDomains.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("account has %d domain(s)\n", len(domains))

	_, _, err = client.SendingDomains.Update(ctx, domain.ID, &mailtrap.UpdateDomainRequest{
		OpenTrackingEnabled:  mailtrap.Ptr(true),
		ClickTrackingEnabled: mailtrap.Ptr(false),
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.SendingDomains.SendSetupInstructions(ctx, domain.ID, "devops@example.com")
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = client.SendingDomains.CreateCompanyInfo(ctx, domain.ID, &mailtrap.CompanyInfoRequest{
		Name:       "Example Inc",
		Address:    "123 Main St",
		City:       "San Francisco",
		Country:    "US",
		ZipCode:    "94105",
		WebsiteURL: "https://example.com",
		InfoLevel:  "business",
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err = client.SendingDomains.Delete(ctx, domain.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted domain")
}
