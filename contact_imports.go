package mailtrap

import (
	"context"
	"fmt"
	"net/http"
)

// ContactImportsService bulk-imports contacts via asynchronous jobs.
type ContactImportsService struct {
	client *Client
}

// Contact import statuses for ContactImport.Status.
const (
	ContactImportCreated  = "created"
	ContactImportStarted  = "started"
	ContactImportFinished = "finished"
	ContactImportFailed   = "failed"
)

// ContactImport is an asynchronous bulk-import job. The count fields are
// populated only once Status is "finished".
type ContactImport struct {
	ID                     int64  `json:"id"`
	Status                 string `json:"status"`
	CreatedContactsCount   int64  `json:"created_contacts_count,omitempty"`
	UpdatedContactsCount   int64  `json:"updated_contacts_count,omitempty"`
	ContactsOverLimitCount int64  `json:"contacts_over_limit_count,omitempty"`
}

// ImportContact is one contact in a bulk import. ListIDsIncluded and
// ListIDsExcluded add and remove list memberships.
type ImportContact struct {
	Email           string         `json:"email"`
	Fields          map[string]any `json:"fields,omitempty"`
	ListIDsIncluded []int64        `json:"list_ids_included,omitempty"`
	ListIDsExcluded []int64        `json:"list_ids_excluded,omitempty"`
}

// Create starts an asynchronous import of the given contacts. Poll Get with the
// returned ID until Status is "finished" or "failed".
func (s *ContactImportsService) Create(ctx context.Context, contacts []*ImportContact) (*ContactImport, *Response, error) {
	body := map[string]any{"contacts": contacts}
	imp := new(ContactImport)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, "/api/contacts/imports", nil, body, imp)
	return imp, resp, err
}

// Get returns the status of a contact import by ID.
func (s *ContactImportsService) Get(ctx context.Context, importID int64) (*ContactImport, *Response, error) {
	path := fmt.Sprintf("/api/contacts/imports/%d", importID)
	imp := new(ContactImport)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, imp)
	return imp, resp, err
}
