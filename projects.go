package mailtrap

import (
	"context"
	"net/http"
)

// ProjectsService manages sandbox projects.
type ProjectsService struct {
	client *Client
}

// ShareLinks holds a project's public share URLs.
type ShareLinks struct {
	Admin  string `json:"admin"`
	Viewer string `json:"viewer"`
}

// Permissions describes the caller's permissions on a resource.
type Permissions struct {
	CanRead    bool `json:"can_read"`
	CanUpdate  bool `json:"can_update"`
	CanDestroy bool `json:"can_destroy"`
	CanLeave   bool `json:"can_leave"`
}

// Project groups sandboxes under a single container.
type Project struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	ShareLinks  *ShareLinks  `json:"share_links,omitempty"`
	Permissions *Permissions `json:"permissions,omitempty"`
}

// List returns all projects accessible to the token.
func (s *ProjectsService) List(ctx context.Context) ([]*Project, *Response, error) {
	path, err := s.client.accountPath("/projects")
	if err != nil {
		return nil, nil, err
	}
	var projects []*Project
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, &projects)
	return projects, resp, err
}

// Get returns a project by ID.
func (s *ProjectsService) Get(ctx context.Context, projectID int64) (*Project, *Response, error) {
	path, err := s.client.accountPath("/projects/%d", projectID)
	if err != nil {
		return nil, nil, err
	}
	project := new(Project)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, path, nil, nil, project)
	return project, resp, err
}

// Create creates a project with the given name.
func (s *ProjectsService) Create(ctx context.Context, name string) (*Project, *Response, error) {
	path, err := s.client.accountPath("/projects")
	if err != nil {
		return nil, nil, err
	}
	body := map[string]any{"project": map[string]string{"name": name}}
	project := new(Project)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPost, path, nil, body, project)
	return project, resp, err
}

// Update renames a project.
func (s *ProjectsService) Update(ctx context.Context, projectID int64, name string) (*Project, *Response, error) {
	path, err := s.client.accountPath("/projects/%d", projectID)
	if err != nil {
		return nil, nil, err
	}
	body := map[string]any{"project": map[string]string{"name": name}}
	project := new(Project)
	resp, err := s.client.do(ctx, HostGeneral, http.MethodPatch, path, nil, body, project)
	return project, resp, err
}

// Delete removes a project by ID.
func (s *ProjectsService) Delete(ctx context.Context, projectID int64) (*Response, error) {
	path, err := s.client.accountPath("/projects/%d", projectID)
	if err != nil {
		return nil, err
	}
	return s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, nil)
}
