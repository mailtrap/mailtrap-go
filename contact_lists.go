package mailtrap

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ContactListsService manages contact lists.
type ContactListsService struct {
	client *Client
}

// ContactList is a named list that groups contacts.
type ContactList struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ContactListListOptions filters a contact list listing.
type ContactListListOptions struct {
	// Search filters lists by name.
	Search string
}

func (o *ContactListListOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	if o.Search != "" {
		v.Set("search", o.Search)
	}
	return v
}

// List returns contact lists matching opts (pass nil for no filters).
func (s *ContactListsService) List(ctx context.Context, opts *ContactListListOptions) ([]*ContactList, *Response, error) {
	var lists []*ContactList
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/contacts/lists", opts.values(), nil, &lists)
	return lists, resp, err
}

// Get returns a contact list by ID.
func (s *ContactListsService) Get(ctx context.Context, listID int64) (*ContactList, *Response, error) {
	path := fmt.Sprintf("/api/contacts/lists/%d", listID)
	list := new(ContactList)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, list)
	return list, resp, err
}

// Create adds a contact list with the given name.
func (s *ContactListsService) Create(ctx context.Context, name string) (*ContactList, *Response, error) {
	body := map[string]string{"name": name}
	list := new(ContactList)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, "/api/contacts/lists", nil, body, list)
	return list, resp, err
}

// Update renames a contact list.
func (s *ContactListsService) Update(ctx context.Context, listID int64, name string) (*ContactList, *Response, error) {
	path := fmt.Sprintf("/api/contacts/lists/%d", listID)
	body := map[string]string{"name": name}
	list := new(ContactList)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPatch, path, nil, body, list)
	return list, resp, err
}

// Delete removes a contact list by ID.
func (s *ContactListsService) Delete(ctx context.Context, listID int64) (*Response, error) {
	path := fmt.Sprintf("/api/contacts/lists/%d", listID)
	return s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, nil)
}
