package mailtrap

import (
	"context"
	"fmt"
	"net/http"
)

// ContactExportsService exports contacts via asynchronous jobs.
type ContactExportsService struct {
	client *Client
}

// Contact export statuses for ContactExport.Status.
const (
	ContactExportCreated  = "created"
	ContactExportStarted  = "started"
	ContactExportFinished = "finished"
)

// Contact export filter names for ContactExportFilter.Name.
const (
	ContactExportFilterListID             = "list_id"
	ContactExportFilterSubscriptionStatus = "subscription_status"
)

// ContactExportOperatorEqual is the only supported ContactExportFilter.Operator.
const ContactExportOperatorEqual = "equal"

// ContactExport is an asynchronous export job. URL is set only once Status is
// "finished".
type ContactExport struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	// URL points to the exported file; nil until the export is finished.
	URL *string `json:"url,omitempty"`
}

// ContactExportFilter narrows an export. Name selects the field (see the
// ContactExportFilter* constants), Operator is "equal", and Value is a []int64
// of list IDs (for list_id) or a subscription-status string.
type ContactExportFilter struct {
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// Create starts an asynchronous export of contacts matching filters (pass nil
// to export all). Poll Get with the returned ID until Status is "finished".
func (s *ContactExportsService) Create(ctx context.Context, filters []*ContactExportFilter) (*ContactExport, *Response, error) {
	body := map[string]any{}
	if len(filters) > 0 {
		body["filters"] = filters
	}
	export := new(ContactExport)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, "/api/contacts/exports", nil, body, export)
	return export, resp, err
}

// Get returns the status of a contact export by ID, including the download URL
// once finished.
func (s *ContactExportsService) Get(ctx context.Context, exportID int64) (*ContactExport, *Response, error) {
	path := fmt.Sprintf("/api/contacts/exports/%d", exportID)
	export := new(ContactExport)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, export)
	return export, resp, err
}
