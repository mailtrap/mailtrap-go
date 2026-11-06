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

	domain, _, err := client.Domains.Create(ctx, "example.com")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created domain %d (%s), status %s\n", domain.ID, domain.DomainName, domain.ComplianceStatus)

	for _, rec := range domain.DNSRecords {
		fmt.Printf("  %s %s %q -> %q (%s)\n", rec.Key, rec.Type, rec.Name, rec.Value, rec.Status)
	}

	domains, _, err := client.Domains.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("account has %d domain(s)\n", len(domains))

	_, _, err = client.Domains.Update(ctx, domain.ID, &mailtrap.UpdateDomainRequest{
		OpenTrackingEnabled:  mailtrap.Ptr(true),
		ClickTrackingEnabled: mailtrap.Ptr(false),
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Domains.SendSetupInstructions(ctx, domain.ID, "devops@example.com")
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = client.Domains.CreateCompanyInfo(ctx, domain.ID, &mailtrap.CompanyInfoRequest{
		Name:       "Example Inc",
		Address:    "123 Main St",
		City:       "San Francisco",
		Country:    "US",
		ZipCode:    "94105",
		WebsiteURL: "https://example.com",
		InfoLevel:  mailtrap.InfoLevelBusiness,
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err = client.Domains.Delete(ctx, domain.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted domain")
}
