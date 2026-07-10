package mailtrap

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// AccountAccessesService lists and removes account accesses: the users,
// invites, and API tokens that hold permissions on an account.
type AccountAccessesService struct {
	client *Client
}

// Account access specifier types for AccountAccess.SpecifierType.
const (
	SpecifierTypeUser     = "User"
	SpecifierTypeInvite   = "Invite"
	SpecifierTypeAPIToken = "ApiToken"
)

// Access levels for AccountAccessResource.AccessLevel, Account.AccessLevels, and
// APITokenPermission.AccessLevel. A higher level grants more rights.
const (
	AccessLevelOwner         = 1000
	AccessLevelAdmin         = 100
	AccessLevelViewerPlus    = 50
	AccessLevelViewer        = 10
	AccessLevelIndeterminate = 1
)

// AccountAccess assigns resource-specific permissions to a specifier (a user,
// invite, or API token).
type AccountAccess struct {
	ID int64 `json:"id"`
	// SpecifierType identifies the kind of entity that holds the access.
	SpecifierType string                    `json:"specifier_type"`
	Specifier     *AccountAccessSpecifier   `json:"specifier"`
	Resources     []*AccountAccessResource  `json:"resources"`
	Permissions   *AccountAccessPermissions `json:"permissions"`
}

// AccountAccessSpecifier describes the entity that holds the access. Which
// fields are set depends on the specifier type: users and invites carry Email,
// while API tokens carry AuthorName, Token, and ExpiresAt.
type AccountAccessSpecifier struct {
	ID                             int64  `json:"id"`
	Email                          string `json:"email,omitempty"`
	Name                           string `json:"name,omitempty"`
	TwoFactorAuthenticationEnabled *bool  `json:"two_factor_authentication_enabled,omitempty"`
	AuthorName                     string `json:"author_name,omitempty"`
	Token                          string `json:"token,omitempty"`
	ExpiresAt                      string `json:"expires_at,omitempty"`
}

// AccountAccessResource is a resource the specifier can access and at what level.
type AccountAccessResource struct {
	ResourceID int64 `json:"resource_id"`
	// ResourceType identifies the kind of resource.
	ResourceType string `json:"resource_type"`
	// AccessLevel is the level granted on the resource.
	AccessLevel int `json:"access_level"`
}

// AccountAccessPermissions reports what the caller may do with an access.
type AccountAccessPermissions struct {
	CanRead    bool `json:"can_read"`
	CanUpdate  bool `json:"can_update"`
	CanDestroy bool `json:"can_destroy"`
	CanLeave   bool `json:"can_leave"`
}

// AccountAccessListOptions filters a listing to accesses on the given
// resources. Leave it nil (or its fields empty) to list all accesses.
type AccountAccessListOptions struct {
	ProjectIDs []int64
	SandboxIDs []int64
	DomainIDs  []int64
}

func (o *AccountAccessListOptions) values() url.Values {
	v := url.Values{}
	if o == nil {
		return v
	}
	for _, id := range o.ProjectIDs {
		v.Add("project_ids[]", strconv.FormatInt(id, 10))
	}
	for _, id := range o.SandboxIDs {
		v.Add("sandbox_ids[]", strconv.FormatInt(id, 10))
	}
	for _, id := range o.DomainIDs {
		v.Add("domain_ids[]", strconv.FormatInt(id, 10))
	}
	return v
}

// List returns account accesses whose specifier is a user or invite, filtered
// by opts (pass nil for no filters). Requires account admin or owner
// permissions.
func (s *AccountAccessesService) List(ctx context.Context, opts *AccountAccessListOptions) ([]*AccountAccess, *Response, error) {
	var accesses []*AccountAccess
	resp, err := s.client.do(ctx, HostGeneral, http.MethodGet, "/api/account_accesses", opts.values(), nil, &accesses)
	return accesses, resp, err
}

// Delete removes an account access by ID. For a user specifier it removes the
// user's permissions; for an invite or token it removes the specifier too.
func (s *AccountAccessesService) Delete(ctx context.Context, accessID int64) (*Response, error) {
	path := fmt.Sprintf("/api/account_accesses/%d", accessID)
	return s.client.do(ctx, HostGeneral, http.MethodDelete, path, nil, nil, nil)
}
