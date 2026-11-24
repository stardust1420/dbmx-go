package dbmx

import (
	"context"
	"net/url"
	"time"
)

// Identity is a user's key identity. Byte fields travel as base64 in JSON.
// EncryptedPrivateKeys is only returned for the caller's own identity.
type Identity struct {
	UserID               string     `json:"user_id"`
	SignPublicKey        []byte     `json:"sign_public_key"`
	BoxPublicKey         []byte     `json:"box_public_key"`
	EncryptedPrivateKeys []byte     `json:"encrypted_private_keys,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	ReplacedAt           *time.Time `json:"replaced_at"`
}

type Device struct {
	ID          string     `json:"id"`
	PublicKey   []byte     `json:"public_key"`
	Name        string     `json:"name"`
	Certificate []byte     `json:"certificate"`
	CreatedAt   time.Time  `json:"created_at"`
	LastSeenAt  *time.Time `json:"last_seen_at"`
}

type ShareGroup struct {
	GroupID          string    `json:"group_id"`
	OwnerUserID      string    `json:"owner_user_id"`
	Audience         string    `json:"audience"`
	DefaultAccess    string    `json:"default_access"`
	DefaultLoginMode string    `json:"default_login_mode"`
	CreatedAt        time.Time `json:"created_at"`
}

// ShareBlob is a share sealed to one device of the recipient.
type ShareBlob struct {
	DeviceID   string `json:"device_id"`
	Ciphertext []byte `json:"ciphertext"`
	Signature  []byte `json:"signature"`
}

type PutShareRequest struct {
	GroupID          string      `json:"group_id"`
	RecipientUserID  string      `json:"recipient_user_id"`
	IncludesPassword bool        `json:"includes_password"`
	Version          int64       `json:"version"`
	Blobs            []ShareBlob `json:"blobs"`
}

type ReceivedShare struct {
	ShareID          string `json:"share_id"`
	GroupID          string `json:"group_id"`
	OrgID            string `json:"org_id"`
	OrgName          string `json:"org_name"`
	SharedByUserID   string `json:"shared_by_user_id"`
	SharedByName     string `json:"shared_by_name"`
	IncludesPassword bool   `json:"includes_password"`
	Version          int64  `json:"version"`
	Ciphertext       []byte `json:"ciphertext"`
	Signature        []byte `json:"signature"`
}

// Missing is a share or a device blob the caller still has to seal.
type Missing struct {
	GroupID         string  `json:"group_id"`
	RecipientUserID string  `json:"recipient_user_id"`
	ShareID         *string `json:"share_id,omitempty"`
	DeviceID        *string `json:"device_id,omitempty"`
	Version         int64   `json:"version,omitempty"`
	NeedsShare      bool    `json:"needs_share"`
}

func (c *Client) GetIdentity(ctx context.Context) (Identity, error) {
	return call[Identity](ctx, c, GET, "/api/v1/identity", nil, "get identity")
}

// PutIdentity sets up or resets the caller's identity.
func (c *Client) PutIdentity(ctx context.Context, signPublicKey, boxPublicKey, encryptedPrivateKeys []byte) error {
	_, err := call[bool](ctx, c, PUT, "/api/v1/identity", map[string][]byte{
		"sign_public_key":        signPublicKey,
		"box_public_key":         boxPublicKey,
		"encrypted_private_keys": encryptedPrivateKeys,
	}, "put identity")
	return err
}

// GetUserIdentity is a co-member's public identity.
func (c *Client) GetUserIdentity(ctx context.Context, userID string) (Identity, error) {
	return call[Identity](ctx, c, GET, "/api/v1/users/"+url.PathEscape(userID)+"/identity", nil, "get user identity")
}

func devicePath(deviceID string, rest ...string) string {
	p := "/api/v1/devices"
	if deviceID != "" {
		p += "/" + url.PathEscape(deviceID)
	}
	for _, r := range rest {
		p += "/" + url.PathEscape(r)
	}
	return p
}

func (c *Client) RegisterDevice(ctx context.Context, publicKey []byte, name string) (Device, error) {
	return call[Device](ctx, c, POST, devicePath(""), struct {
		PublicKey []byte `json:"public_key"`
		Name      string `json:"name"`
	}{publicKey, name}, "register device")
}

func (c *Client) ListDevices(ctx context.Context) ([]Device, error) {
	return call[[]Device](ctx, c, GET, devicePath(""), nil, "list devices")
}

// CertifyDevice is sent by a device that unlocked the identity with the
// recovery code, certifying itself.
func (c *Client) CertifyDevice(ctx context.Context, deviceID string, certificate []byte) error {
	_, err := call[bool](ctx, c, POST, devicePath(deviceID, "certify"),
		map[string][]byte{"certificate": certificate}, "certify device")
	return err
}

// ApproveDevice certifies a pending device and seals the identity to it.
func (c *Client) ApproveDevice(ctx context.Context, deviceID string, certificate, encryptedIdentity []byte) error {
	_, err := call[bool](ctx, c, POST, devicePath(deviceID, "approve"), map[string][]byte{
		"certificate":        certificate,
		"encrypted_identity": encryptedIdentity,
	}, "approve device")
	return err
}

// GetPendingDeviceIdentity returns the identity sealed to a pending device.
// The server answers 404 (an *APIError) until the device is approved.
func (c *Client) GetPendingDeviceIdentity(ctx context.Context, deviceID string) ([]byte, error) {
	out, err := call[struct {
		EncryptedIdentity []byte `json:"encrypted_identity"`
	}](ctx, c, GET, devicePath(deviceID, "identity"), nil, "get pending device identity")
	return out.EncryptedIdentity, err
}

func (c *Client) RevokeDevice(ctx context.Context, deviceID string) error {
	_, err := call[bool](ctx, c, DELETE, devicePath(deviceID), nil, "revoke device")
	return err
}

func (c *Client) PutShare(ctx context.Context, orgID string, req PutShareRequest) error {
	_, err := call[bool](ctx, c, PUT, orgPath(orgID, "shares"), req, "put share")
	return err
}

// DeleteShare deletes one recipient's share, or the whole group when
// recipientUserID is empty.
func (c *Client) DeleteShare(ctx context.Context, orgID, groupID, recipientUserID string) error {
	q := url.Values{"group_id": {groupID}}
	if recipientUserID != "" {
		q.Set("recipient_user_id", recipientUserID)
	}
	_, err := call[bool](ctx, c, DELETE, orgPath(orgID, "shares")+"?"+q.Encode(), nil, "delete share")
	return err
}

func (c *Client) PutShareGroup(ctx context.Context, orgID string, g ShareGroup) error {
	_, err := call[bool](ctx, c, PUT, orgPath(orgID, "share-groups"), g, "put share group")
	return err
}

func (c *Client) ListShareGroups(ctx context.Context, orgID string) ([]ShareGroup, error) {
	return call[[]ShareGroup](ctx, c, GET, orgPath(orgID, "share-groups"), nil, "list share groups")
}

// ListReceivedShares lists the shares sealed to the caller's device.
func (c *Client) ListReceivedShares(ctx context.Context, deviceID string) ([]ReceivedShare, error) {
	return callWithHeaders[[]ReceivedShare](ctx, c, GET, "/api/v1/shares", nil,
		map[string]string{"X-Device-ID": deviceID}, "list received shares")
}

// ListMissing lists the shares and device blobs the caller still has to seal.
func (c *Client) ListMissing(ctx context.Context, orgID string) ([]Missing, error) {
	return call[[]Missing](ctx, c, GET, orgPath(orgID, "shares", "missing"), nil, "list missing shares")
}
