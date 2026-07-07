package mailtrap

import (
	"context"
	"fmt"
	"net/http"
)

// PermissionsService lists the resources a token can administer and updates the
// permissions attached to an account access.
type PermissionsService struct {
	client *Client
}

// Resource types for the resource_type fields of PermissionUpdate,
// AccountAccessResource, and APITokenPermission, and for PermissionResource.Type.
const (
	ResourceTypeAccount            = "account"
	ResourceTypeBilling            = "billing"
	ResourceTypeOrganization       = "organization"
	ResourceTypeProject            = "project"
	ResourceTypeSandbox            = "sandbox"
	ResourceTypeDomain             = "domain"
	ResourceTypeEmailCampaignScope = "email_campaign_permission_scope"
	ResourceTypeEmailTemplateScope = "email_template_permission_scope"
)

// Permission access levels for PermissionUpdate.AccessLevel. The bulk-update
// endpoint also accepts the numeric equivalents ("100", "10").
const (
	PermissionLevelAdmin  = "admin"
	PermissionLevelViewer = "viewer"
)

// PermissionResource is a node in the account's resource hierarchy (the account,
// projects, sandboxes, domains, and so on) that the token can administer.
type PermissionResource struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// Type identifies the kind of resource.
	Type string `json:"type"`
	// AccessLevel is the token's access level for this resource.
	AccessLevel int                   `json:"access_level"`
	Resources   []*PermissionResource `json:"resources"`
}

// PermissionUpdate creates, updates, or removes one permission in a bulk update.
// A resource_type/resource_id pair that already exists is updated; otherwise it
// is created. Set Destroy to remove it instead.
type PermissionUpdate struct {
	ResourceID string `json:"resource_id"`
	// ResourceType identifies the kind of resource.
	ResourceType string `json:"resource_type"`
	// AccessLevel is the level to grant; leave empty when Destroy is true.
	AccessLevel string `json:"access_level,omitempty"`
	// Destroy removes the permission instead of creating or updating it.
	Destroy bool `json:"_destroy,omitempty"`
}

// Resources returns the account resources the token has admin access to, nested
// by hierarchy.
func (s *PermissionsService) Resources(ctx context.Context) ([]*PermissionResource, *Response, error) {
	var resources []*PermissionResource
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/permissions/resources", nil, nil, &resources)
	return resources, resp, err
}

// BulkUpdate creates, updates, or removes permissions for an account access in
// a single request.
func (s *PermissionsService) BulkUpdate(ctx context.Context, accessID int64, permissions []*PermissionUpdate) (*Response, error) {
	path := fmt.Sprintf("/api/account_accesses/%d/permissions/bulk", accessID)
	body := map[string]any{"permissions": permissions}
	return s.client.do(ctx, HostGeneral, http.MethodPut, path, nil, body, nil)
}
