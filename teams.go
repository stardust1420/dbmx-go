package dbmx

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

type Org struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OwnerUserID string    `json:"owner_user_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// OrgSummary is an org as listed for one user: their role and the headcount.
type OrgSummary struct {
	Org
	Role        string `json:"role"`
	MemberCount int    `json:"member_count"`
}

type Member struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// Invite is an invite as the org's owner sees it.
type Invite struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Email     string    `json:"email"`
	InvitedBy string    `json:"invited_by"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PendingInvite is an invite addressed to the signed-in user. The server
// returns the inviter's ID, not their name.
type PendingInvite struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	OrgName   string    `json:"org_name"`
	InvitedBy string    `json:"invited_by"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OrgBilling is an org's Team plan, in the shape of Billing plus its seats.
type OrgBilling struct {
	Billing
	SeatCount          int     `json:"seat_count"`
	SeatsOutOfSync     bool    `json:"seats_out_of_sync"`
	TrialMemberCap     int     `json:"trial_member_cap"`
	PendingOwnerUserID *string `json:"pending_owner_user_id"`
}

type orgIDReq struct {
	OrgID string `json:"org_id"`
}

// call sends one request and returns the envelope's data.
func call[T any](ctx context.Context, c *Client, method httpMethod, path string, payload any, what string) (T, error) {
	return callWithHeaders[T](ctx, c, method, path, payload, nil, what)
}

func callWithHeaders[T any](ctx context.Context, c *Client, method httpMethod, path string, payload any, headers map[string]string, what string) (T, error) {
	var zero T
	res, err := httpHandler(ctx, httpHandlerArgs{URL: path, Method: method, Payload: payload, Credentials: c.Credentials, Headers: headers})
	if err != nil {
		return zero, errors.Wrap(err, "Failed to "+what)
	}
	var out struct {
		Data T `json:"data"`
	}
	if err := json.Unmarshal(res, &out); err != nil {
		return zero, errors.Wrap(err, "Failed to unmarshal response to "+what)
	}
	return out.Data, nil
}

func orgPath(orgID string, rest ...string) string {
	p := "/api/v1/orgs/" + url.PathEscape(orgID)
	for _, r := range rest {
		p += "/" + url.PathEscape(r)
	}
	return p
}

func (c *Client) CreateOrg(ctx context.Context, name string) (Org, error) {
	return call[Org](ctx, c, POST, "/api/v1/orgs", map[string]string{"name": name}, "create org")
}

func (c *Client) ListOrgs(ctx context.Context) ([]OrgSummary, error) {
	return call[[]OrgSummary](ctx, c, GET, "/api/v1/orgs", nil, "list orgs")
}

func (c *Client) GetOrg(ctx context.Context, orgID string) (Org, error) {
	return call[Org](ctx, c, GET, orgPath(orgID), nil, "get org")
}

func (c *Client) RenameOrg(ctx context.Context, orgID, name string) error {
	_, err := call[bool](ctx, c, "PATCH", orgPath(orgID), map[string]string{"name": name}, "rename org")
	return err
}

func (c *Client) DeleteOrg(ctx context.Context, orgID string) error {
	_, err := call[bool](ctx, c, DELETE, orgPath(orgID), nil, "delete org")
	return err
}

func (c *Client) ListMembers(ctx context.Context, orgID string) ([]Member, error) {
	return call[[]Member](ctx, c, GET, orgPath(orgID, "members"), nil, "list members")
}

func (c *Client) RemoveMember(ctx context.Context, orgID, userID string) error {
	_, err := call[bool](ctx, c, DELETE, orgPath(orgID, "members", userID), nil, "remove member")
	return err
}

// TransferOrg hands the org to another member. pending is true when it waits
// for them to subscribe.
func (c *Client) TransferOrg(ctx context.Context, orgID, userID string) (pending bool, err error) {
	out, err := call[struct {
		Pending bool `json:"pending"`
	}](ctx, c, POST, orgPath(orgID, "transfer"), map[string]string{"user_id": userID}, "transfer org")
	return out.Pending, err
}

func (c *Client) InviteMember(ctx context.Context, orgID, email string) (Invite, error) {
	return call[Invite](ctx, c, POST, orgPath(orgID, "invites"), map[string]string{"email": email}, "invite member")
}

func (c *Client) ListOrgInvites(ctx context.Context, orgID string) ([]Invite, error) {
	return call[[]Invite](ctx, c, GET, orgPath(orgID, "invites"), nil, "list invites")
}

func (c *Client) ResendInvite(ctx context.Context, orgID, inviteID string) (Invite, error) {
	return call[Invite](ctx, c, POST, orgPath(orgID, "invites", inviteID, "resend"), nil, "resend invite")
}

func (c *Client) RevokeInvite(ctx context.Context, orgID, inviteID string) error {
	_, err := call[bool](ctx, c, DELETE, orgPath(orgID, "invites", inviteID), nil, "revoke invite")
	return err
}

func (c *Client) ListMyInvites(ctx context.Context) ([]PendingInvite, error) {
	return call[[]PendingInvite](ctx, c, GET, "/api/v1/invites", nil, "list invites")
}

func (c *Client) AcceptInvite(ctx context.Context, inviteID string) (Org, error) {
	return call[Org](ctx, c, POST, "/api/v1/invites/"+url.PathEscape(inviteID)+"/accept", nil, "accept invite")
}

func (c *Client) AcceptInviteToken(ctx context.Context, token string) (Org, error) {
	return call[Org](ctx, c, POST, "/api/v1/invites/accept-token", map[string]string{"token": token}, "accept invite")
}

func (c *Client) StartTeamTrial(ctx context.Context, orgID string) (OrgBilling, error) {
	return call[OrgBilling](ctx, c, POST, "/api/v1/billing/team/trial", orgIDReq{OrgID: orgID}, "start team trial")
}

func (c *Client) CreateTeamCheckout(ctx context.Context, orgID string) (string, error) {
	out, err := call[struct {
		URL string `json:"url"`
	}](ctx, c, POST, "/api/v1/billing/team/checkout", orgIDReq{OrgID: orgID}, "create team checkout")
	return out.URL, err
}

func (c *Client) GetOrgBilling(ctx context.Context, orgID string) (OrgBilling, error) {
	return call[OrgBilling](ctx, c, GET, orgPath(orgID, "billing"), nil, "get org billing")
}

func (c *Client) CreateOrgPortal(ctx context.Context, orgID string) (string, error) {
	out, err := call[struct {
		URL string `json:"url"`
	}](ctx, c, POST, orgPath(orgID, "billing", "portal"), nil, "create org portal")
	return out.URL, err
}
