package mailtrap

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

// SubscriptionTopicsService reads the account's subscription topics: the
// kinds of email marketing messages a recipient can opt out of one by one.
// Topics are managed in the Mailtrap app; the API only reads them, so their
// IDs can be set on campaigns and contacts.
//
// On an account without subscription topics, every call fails with a
// ForbiddenError.
type SubscriptionTopicsService struct {
	client *Client
}

// SubscriptionTopic is a subscription topic.
type SubscriptionTopic struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	// DomainID is the ID of the domain the topic belongs to, as returned by
	// the Domains endpoints, or nil when the topic covers every domain.
	DomainID *int64 `json:"domain_id"`
	// DeletedAt is set on a deleted topic, which only Get returns.
	DeletedAt *string `json:"deleted_at"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// SubscriptionTopicListOptions paginates a subscription topic listing.
type SubscriptionTopicListOptions struct {
	// Token is the page number to retrieve (page-token pagination); zero means
	// the first page.
	Token int
	// PerPage is the number of topics per page (default 50, maximum 100).
	PerPage int
}

func (o *SubscriptionTopicListOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	if o.Token > 0 {
		v.Set("token", strconv.Itoa(o.Token))
	}
	if o.PerPage > 0 {
		v.Set("per_page", strconv.Itoa(o.PerPage))
	}
	return v
}

// SubscriptionTopicsList is a page of subscription topics with pagination
// metadata.
type SubscriptionTopicsList struct {
	Data       []*SubscriptionTopic `json:"data"`
	Pagination Pagination           `json:"pagination"`
}

// List returns a page of the account's subscription topics, ordered by ID
// (pass nil for the first page with defaults). Deleted topics are left out.
// Follow Pagination.NextToken with Token for the next page, or use All to
// iterate every topic.
func (s *SubscriptionTopicsService) List(ctx context.Context, opts *SubscriptionTopicListOptions) (*SubscriptionTopicsList, *Response, error) {
	list := new(SubscriptionTopicsList)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/subscription_topics", opts.values(), nil, list)
	return list, resp, err
}

// All iterates every subscription topic, following the page token across
// pages. Iteration stops at the first error, yielded once.
func (s *SubscriptionTopicsService) All(ctx context.Context, opts *SubscriptionTopicListOptions) iter.Seq2[*SubscriptionTopic, error] {
	return func(yield func(*SubscriptionTopic, error) bool) {
		o := SubscriptionTopicListOptions{}
		if opts != nil {
			o = *opts
		}
		for {
			list, _, err := s.List(ctx, &o)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, t := range list.Data {
				if !yield(t, nil) {
					return
				}
			}
			if list.Pagination.NextToken == nil {
				return
			}
			o.Token = *list.Pagination.NextToken
		}
	}
}

// Get returns a subscription topic by ID, including a deleted one (with
// DeletedAt set), so every topic ID a campaign carries resolves.
func (s *SubscriptionTopicsService) Get(ctx context.Context, subscriptionTopicID int64) (*SubscriptionTopic, *Response, error) {
	path := fmt.Sprintf("/api/subscription_topics/%d", subscriptionTopicID)
	var wrapper struct {
		Data *SubscriptionTopic `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, &wrapper)
	return wrapper.Data, resp, err
}
