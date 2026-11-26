package mailtrap_test

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/mailtrap/mailtrap-go"
)

func TestContacts_Create(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("POST /api/contacts", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"contact":{"email":"john@example.com","fields":{"first_name":"John"},"list_ids":[1,2],`+
			`"subscription_topics":[{"subscription_topic_id":12,"domain_id":4321,"status":"unsubscribed"}]}}`)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":"018dd5e3","email":"john@example.com","status":"subscribed","list_ids":[1,2],` +
			`"subscription_topics":[{"subscription_topic_id":12,"domain_id":4321,"status":"unsubscribed"},` +
			`{"subscription_topic_id":12,"domain_id":4322,"status":"subscribed"}]}}`))
	})

	contact, _, err := client.Contacts.Create(context.Background(), &mailtrap.CreateContactRequest{
		Email:   "john@example.com",
		Fields:  map[string]any{"first_name": "John"},
		ListIDs: []int64{1, 2},
		SubscriptionTopics: []mailtrap.ContactSubscriptionTopic{
			{SubscriptionTopicID: 12, DomainID: 4321, Status: mailtrap.ContactStatusUnsubscribed},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if contact.ID != "018dd5e3" || contact.Status != mailtrap.ContactStatusSubscribed {
		t.Errorf("contact = %+v", contact)
	}
	want := []mailtrap.ContactSubscriptionTopic{
		{SubscriptionTopicID: 12, DomainID: 4321, Status: mailtrap.ContactStatusUnsubscribed},
		{SubscriptionTopicID: 12, DomainID: 4322, Status: mailtrap.ContactStatusSubscribed},
	}
	if !reflect.DeepEqual(contact.SubscriptionTopics, want) {
		t.Errorf("SubscriptionTopics = %+v, want %+v", contact.SubscriptionTopics, want)
	}
}

func TestContacts_Create_subscriptionTopicsNotSaved(t *testing.T) {
	mux, client := setup(t)
	const msg = "The contact was saved, but its subscription topic preferences were not. " +
		"Send them again in an update of contact 018dd5e3."
	mux.HandleFunc("POST /api/contacts", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"errors":"` + msg + `"}`))
	})

	_, _, err := client.Contacts.Create(context.Background(), &mailtrap.CreateContactRequest{
		Email: "john@example.com",
		SubscriptionTopics: []mailtrap.ContactSubscriptionTopic{
			{SubscriptionTopicID: 12, DomainID: 4321, Status: mailtrap.ContactStatusUnsubscribed},
		},
	})
	var apiErr *mailtrap.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if apiErr.StatusCode != http.StatusServiceUnavailable || len(apiErr.Messages) != 1 || apiErr.Messages[0] != msg {
		t.Errorf("err = %+v", apiErr)
	}
}

func TestContacts_Get(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/018dd5e3", func(w http.ResponseWriter, _ *http.Request) {
		// null: the topic preferences could not be read at the moment.
		_, _ = w.Write([]byte(`{"data":{"id":"018dd5e3","email":"john@example.com","status":"unsubscribed","subscription_topics":null}}`))
	})

	contact, _, err := client.Contacts.Get(context.Background(), "018dd5e3")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if contact.Email != "john@example.com" || contact.Status != mailtrap.ContactStatusUnsubscribed {
		t.Errorf("contact = %+v", contact)
	}
	if contact.SubscriptionTopics != nil {
		t.Errorf("SubscriptionTopics = %+v, want nil", contact.SubscriptionTopics)
	}
}

func TestContacts_Get_byEmail(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("GET /api/contacts/john@example.com", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"id":"018dd5e3","email":"john@example.com"}}`))
	})

	contact, _, err := client.Contacts.Get(context.Background(), "john@example.com")
	if err != nil {
		t.Fatalf("Get by email: %v", err)
	}
	// No subscription_topics key: the account has no subscription topics.
	if contact.ID != "018dd5e3" || contact.SubscriptionTopics != nil {
		t.Errorf("contact = %+v", contact)
	}
}

func TestContacts_Update(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/contacts/018dd5e3", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"contact":{"email":"john@example.com","list_ids_included":[3],"list_ids_excluded":[1],"unsubscribed":true}}`)
		_, _ = w.Write([]byte(`{"action":"updated","data":{"id":"018dd5e3","email":"john@example.com","status":"unsubscribed"}}`))
	})

	upsert, _, err := client.Contacts.Update(context.Background(), "018dd5e3", &mailtrap.UpdateContactRequest{
		Email:           "john@example.com",
		ListIDsIncluded: []int64{3},
		ListIDsExcluded: []int64{1},
		Unsubscribed:    mailtrap.Ptr(true),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if upsert.Action != "updated" || upsert.Contact.ID != "018dd5e3" {
		t.Errorf("upsert = %+v (contact %+v)", upsert, upsert.Contact)
	}
}

func TestContacts_Update_subscriptionTopicsOnly(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/contacts/018dd5e3", func(w http.ResponseWriter, r *http.Request) {
		wantJSONBody(t, r, `{"contact":{"email":"john@example.com","subscription_topics":[{"subscription_topic_id":12,"domain_id":4321,"status":"subscribed"}]}}`)
		_, _ = w.Write([]byte(`{"action":"updated","data":{"id":"018dd5e3","email":"john@example.com","status":"subscribed",` +
			`"subscription_topics":[{"subscription_topic_id":12,"domain_id":4321,"status":"subscribed"}]}}`))
	})

	upsert, _, err := client.Contacts.Update(context.Background(), "018dd5e3", &mailtrap.UpdateContactRequest{
		Email: "john@example.com",
		SubscriptionTopics: []mailtrap.ContactSubscriptionTopic{
			{SubscriptionTopicID: 12, DomainID: 4321, Status: mailtrap.ContactStatusSubscribed},
		},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	want := []mailtrap.ContactSubscriptionTopic{
		{SubscriptionTopicID: 12, DomainID: 4321, Status: mailtrap.ContactStatusSubscribed},
	}
	if !reflect.DeepEqual(upsert.Contact.SubscriptionTopics, want) {
		t.Errorf("SubscriptionTopics = %+v, want %+v", upsert.Contact.SubscriptionTopics, want)
	}
}

func TestContacts_Update_subscriptionTopicsUnavailable(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("PATCH /api/contacts/018dd5e3", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"errors":{"subscription_topics":["are not available for this account"]}}`))
	})

	_, _, err := client.Contacts.Update(context.Background(), "018dd5e3", &mailtrap.UpdateContactRequest{
		Email: "john@example.com",
		SubscriptionTopics: []mailtrap.ContactSubscriptionTopic{
			{SubscriptionTopicID: 12, DomainID: 4321, Status: mailtrap.ContactStatusSubscribed},
		},
	})
	var validation *mailtrap.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("err = %v, want *ValidationError", err)
	}
	if got := validation.Fields["subscription_topics"]; !reflect.DeepEqual(got, []string{"are not available for this account"}) {
		t.Errorf("Fields[subscription_topics] = %v", got)
	}
}

func TestContacts_Delete(t *testing.T) {
	mux, client := setup(t)
	mux.HandleFunc("DELETE /api/contacts/018dd5e3", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	resp, err := client.Contacts.Delete(context.Background(), "018dd5e3")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d", resp.StatusCode)
	}
}
