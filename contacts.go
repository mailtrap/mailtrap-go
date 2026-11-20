package mailtrap

import (
	"context"
	"net/http"
	"net/url"
)

// ContactsService manages email marketing contacts.
type ContactsService struct {
	client *Client
}

// Contact subscription statuses for Contact.Status.
const (
	ContactStatusSubscribed   = "subscribed"
	ContactStatusUnsubscribed = "unsubscribed"
)

// Contact is an email marketing contact.
type Contact struct {
	// ID is the contact's UUID.
	ID    string `json:"id"`
	Email string `json:"email"`
	// Fields maps custom field merge tags to scalar values.
	Fields  map[string]any `json:"fields,omitempty"`
	ListIDs []int64        `json:"list_ids,omitempty"`
	// Status is the subscription status.
	Status string `json:"status,omitempty"`
	// CreatedAt and UpdatedAt are Unix timestamps in milliseconds.
	CreatedAt int64 `json:"created_at,omitempty"`
	UpdatedAt int64 `json:"updated_at,omitempty"`
	// SubscriptionTopics is the contact's status for each subscription topic
	// on each domain. Get, Create and Update return it on an account with
	// subscription topics. It is nil when the API leaves it out (an account
	// without subscription topics) and when the preferences could not be read
	// at the moment (null; retry later).
	SubscriptionTopics []ContactSubscriptionTopic `json:"subscription_topics,omitempty"`
}

// ContactSubscriptionTopic is a contact's status for one subscription topic
// on one domain.
type ContactSubscriptionTopic struct {
	// SubscriptionTopicID is the topic's ID, as returned by
	// SubscriptionTopicsService.
	SubscriptionTopicID int64 `json:"subscription_topic_id"`
	// DomainID is the ID of the domain the status applies to, as returned by
	// the Domains endpoints.
	DomainID int64 `json:"domain_id"`
	// Status is ContactStatusSubscribed or ContactStatusUnsubscribed.
	Status string `json:"status"`
}

// CreateContactRequest is the payload for creating a contact. Email is required.
type CreateContactRequest struct {
	Email   string         `json:"email"`
	Fields  map[string]any `json:"fields,omitempty"`
	ListIDs []int64        `json:"list_ids,omitempty"`
	// SubscriptionTopics sets the contact's status for these subscription
	// topics on these domains. On an account without subscription topics,
	// setting it fails with a ValidationError.
	SubscriptionTopics []ContactSubscriptionTopic `json:"subscription_topics,omitempty"`
}

// UpdateContactRequest is the payload for updating (upserting) a contact. Email
// is required; ListIDsIncluded and ListIDsExcluded add and remove list
// memberships.
type UpdateContactRequest struct {
	Email           string         `json:"email"`
	Fields          map[string]any `json:"fields,omitempty"`
	ListIDsIncluded []int64        `json:"list_ids_included,omitempty"`
	ListIDsExcluded []int64        `json:"list_ids_excluded,omitempty"`
	Unsubscribed    *bool          `json:"unsubscribed,omitempty"`
	// SubscriptionTopics sets the contact's status for these subscription
	// topics on these domains; the topic and domain pairs left out keep their
	// status. On an account without subscription topics, setting it fails with
	// a ValidationError.
	SubscriptionTopics []ContactSubscriptionTopic `json:"subscription_topics,omitempty"`
}

// ContactUpsert is the result of Update: the contact plus whether it was
// created or updated.
type ContactUpsert struct {
	// Action is "created" or "updated".
	Action  string   `json:"action"`
	Contact *Contact `json:"data"`
}

// Create adds a contact.
//
// When the contact is saved but its SubscriptionTopics are not, Create fails
// with an *Error of status 503 whose message names the contact's ID: send the
// topics again with Update.
func (s *ContactsService) Create(ctx context.Context, req *CreateContactRequest) (*Contact, *Response, error) {
	var wrapper struct {
		Data *Contact `json:"data"`
	}
	body := map[string]any{"contact": req}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, "/api/contacts", nil, body, &wrapper)
	return wrapper.Data, resp, err
}

// Get returns a contact by UUID or email.
func (s *ContactsService) Get(ctx context.Context, identifier string) (*Contact, *Response, error) {
	var wrapper struct {
		Data *Contact `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, contactPath(identifier), nil, nil, &wrapper)
	return wrapper.Data, resp, err
}

// Update creates or updates (upserts) the contact identified by UUID or email,
// reporting which action was taken.
//
// When the contact is saved but its SubscriptionTopics are not, Update fails
// with an *Error of status 503 whose message names the contact's ID: send the
// topics again in an Update of that ID.
func (s *ContactsService) Update(ctx context.Context, identifier string, req *UpdateContactRequest) (*ContactUpsert, *Response, error) {
	upsert := new(ContactUpsert)
	body := map[string]any{"contact": req}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPatch, contactPath(identifier), nil, body, upsert)
	return upsert, resp, err
}

// Delete removes a contact by UUID or email.
func (s *ContactsService) Delete(ctx context.Context, identifier string) (*Response, error) {
	return s.client.do(ctx, HostGeneral, http.MethodDelete, contactPath(identifier), nil, nil, nil)
}

// contactPath builds the path for a contact identifier (UUID or email),
// escaping it so an email address can be passed unencoded.
func contactPath(identifier string) string {
	return "/api/contacts/" + url.PathEscape(identifier)
}
