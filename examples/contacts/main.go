package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"

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

	// Unsubscribe the contact from one subscription topic on one domain; the
	// topics and domains left out keep their status.
	if topicID, domainID, ok := topicOnDomain(ctx, client); ok {
		upsert, _, err := client.Contacts.Update(ctx, contact.ID, &mailtrap.UpdateContactRequest{
			Email: "john.smith@example.com",
			SubscriptionTopics: []mailtrap.ContactSubscriptionTopic{
				{SubscriptionTopicID: topicID, DomainID: domainID, Status: mailtrap.ContactStatusUnsubscribed},
			},
		})
		var validation *mailtrap.ValidationError
		switch {
		case errors.As(err, &validation) && slices.Contains(validation.Fields["subscription_topics"], "are not available for this account"):
			// An account can have topics before it can set them on contacts.
			// Other subscription_topics errors are real problems with the request.
			fmt.Println("contact topic preferences are not available for this account")
		case err != nil:
			log.Fatal(err)
		default:
			for _, t := range upsert.Contact.SubscriptionTopics {
				fmt.Printf("topic %d on domain %d: %s\n", t.SubscriptionTopicID, t.DomainID, t.Status)
			}
		}
	}

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

// topicOnDomain picks the first subscription topic and a domain it covers, so
// the example needs no hardcoded IDs. ok is false when the account has no
// subscription topics.
func topicOnDomain(ctx context.Context, client *mailtrap.Client) (topicID, domainID int64, ok bool) {
	page, _, err := client.SubscriptionTopics.List(ctx, &mailtrap.SubscriptionTopicListOptions{PerPage: 1})
	var forbidden *mailtrap.ForbiddenError
	if errors.As(err, &forbidden) {
		fmt.Println("subscription topics are not available for this account")
		return 0, 0, false
	}
	if err != nil {
		log.Fatal(err)
	}
	if len(page.Data) == 0 {
		return 0, 0, false
	}
	topic := page.Data[0]
	if topic.DomainID != nil {
		return topic.ID, *topic.DomainID, true
	}

	// A topic for every domain has a status on each of them; take the first.
	domains, _, err := client.Domains.List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if len(domains) == 0 {
		return 0, 0, false
	}
	return topic.ID, domains[0].ID, true
}
