package mailtrap

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strconv"
)

// TemplatesService manages the account's email templates through the
// paginated /api/templates endpoints.
type TemplatesService struct {
	client *Client
}

// Template is a reusable email template.
type Template struct {
	ID        int64  `json:"id"`
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Subject   string `json:"subject"`
	BodyText  string `json:"body_text,omitempty"`
	BodyHTML  string `json:"body_html,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateTemplateRequest is the payload for creating a template. Name, Subject,
// and Category are required; the bodies are optional.
type CreateTemplateRequest struct {
	Name     string `json:"name"`
	Subject  string `json:"subject"`
	Category string `json:"category"`
	BodyText string `json:"body_text,omitempty"`
	BodyHTML string `json:"body_html,omitempty"`
}

// UpdateTemplateRequest changes a template. All fields are optional; only the
// set fields change.
type UpdateTemplateRequest struct {
	Name     string `json:"name,omitempty"`
	Subject  string `json:"subject,omitempty"`
	Category string `json:"category,omitempty"`
	BodyText string `json:"body_text,omitempty"`
	BodyHTML string `json:"body_html,omitempty"`
}

// TemplateListOptions paginates a template listing.
type TemplateListOptions struct {
	// Token is the page number to retrieve (page-token pagination); zero means
	// the first page.
	Token int
	// PerPage is the number of templates per page (default 50, maximum 100).
	PerPage int
}

func (o *TemplateListOptions) values() url.Values {
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

// TemplatesList is a page of templates with pagination metadata.
type TemplatesList struct {
	Data       []*Template `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// List returns a page of templates (pass nil opts for the first page with
// defaults). Follow Pagination.NextToken with Token for the next page, or use
// All to iterate every template.
func (s *TemplatesService) List(ctx context.Context, opts *TemplateListOptions) (*TemplatesList, *Response, error) {
	list := new(TemplatesList)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/templates", opts.values(), nil, list)
	return list, resp, err
}

// All iterates every template, following the page token across pages.
// Iteration stops at the first error, yielded once.
func (s *TemplatesService) All(ctx context.Context, opts *TemplateListOptions) iter.Seq2[*Template, error] {
	return func(yield func(*Template, error) bool) {
		o := TemplateListOptions{}
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

// Get returns a template by ID.
func (s *TemplatesService) Get(ctx context.Context, templateID int64) (*Template, *Response, error) {
	path := fmt.Sprintf("/api/templates/%d", templateID)
	return s.doTemplate(ctx, http.MethodGet, path, nil)
}

// Create adds a template.
func (s *TemplatesService) Create(ctx context.Context, req *CreateTemplateRequest) (*Template, *Response, error) {
	return s.doTemplate(ctx, http.MethodPost, "/api/templates", req)
}

// Update changes the set fields of a template.
func (s *TemplatesService) Update(ctx context.Context, templateID int64, req *UpdateTemplateRequest) (*Template, *Response, error) {
	path := fmt.Sprintf("/api/templates/%d", templateID)
	return s.doTemplate(ctx, http.MethodPatch, path, req)
}

// Delete removes a template by ID.
func (s *TemplatesService) Delete(ctx context.Context, templateID int64) (*Response, error) {
	path := fmt.Sprintf("/api/templates/%d", templateID)
	return s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, nil)
}

// doTemplate sends a request and unwraps the single-template data envelope
// shared by Get, Create, and Update.
func (s *TemplatesService) doTemplate(ctx context.Context, method, path string, body any) (*Template, *Response, error) {
	var wrapper struct {
		Data *Template `json:"data"`
	}
	resp, err := s.client.do(ctx, HostGeneral, method, path, nil, body, &wrapper)
	return wrapper.Data, resp, err
}
