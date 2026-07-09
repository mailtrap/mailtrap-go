package mailtrap

import (
	"context"
	"fmt"
	"net/http"
)

// ContactFieldsService manages custom contact fields.
type ContactFieldsService struct {
	client *Client
}

// Contact field data types for ContactField.DataType.
const (
	ContactFieldTypeText    = "text"
	ContactFieldTypeInteger = "integer"
	ContactFieldTypeFloat   = "float"
	ContactFieldTypeBoolean = "boolean"
	ContactFieldTypeDate    = "date"
)

// ContactField is a custom field that can be set on contacts.
type ContactField struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// DataType is the field's value type.
	DataType string `json:"data_type"`
	// MergeTag personalizes campaigns with the field's per-contact value.
	MergeTag string `json:"merge_tag"`
}

// CreateContactFieldRequest is the payload for creating a contact field. All
// fields are required.
type CreateContactFieldRequest struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	MergeTag string `json:"merge_tag"`
}

// UpdateContactFieldRequest changes a contact field's name and merge tag; the
// data type is immutable.
type UpdateContactFieldRequest struct {
	Name     string `json:"name"`
	MergeTag string `json:"merge_tag"`
}

// List returns all contact fields.
func (s *ContactFieldsService) List(ctx context.Context) ([]*ContactField, *Response, error) {
	var fields []*ContactField
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/contacts/fields", nil, nil, &fields)
	return fields, resp, err
}

// Get returns a contact field by ID.
func (s *ContactFieldsService) Get(ctx context.Context, fieldID int64) (*ContactField, *Response, error) {
	path := fmt.Sprintf("/api/contacts/fields/%d", fieldID)
	field := new(ContactField)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, field)
	return field, resp, err
}

// Create adds a contact field.
func (s *ContactFieldsService) Create(ctx context.Context, req *CreateContactFieldRequest) (*ContactField, *Response, error) {
	field := new(ContactField)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, "/api/contacts/fields", nil, req, field)
	return field, resp, err
}

// Update changes a contact field's name and merge tag.
func (s *ContactFieldsService) Update(ctx context.Context, fieldID int64, req *UpdateContactFieldRequest) (*ContactField, *Response, error) {
	path := fmt.Sprintf("/api/contacts/fields/%d", fieldID)
	field := new(ContactField)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPatch, path, nil, req, field)
	return field, resp, err
}

// Delete removes a contact field by ID.
func (s *ContactFieldsService) Delete(ctx context.Context, fieldID int64) (*Response, error) {
	path := fmt.Sprintf("/api/contacts/fields/%d", fieldID)
	return s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, nil)
}
