package dbmx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	Credentials
}

type Credentials struct {
	BaseURL     string
	AccessToken string
}

func NewClient(cr Credentials) Client {
	return Client{
		Credentials: Credentials{
			BaseURL:     strings.TrimSpace(cr.BaseURL),
			AccessToken: strings.TrimSpace(cr.AccessToken),
		},
	}
}

// Http helper
type httpMethod string

const (
	GET    httpMethod = "GET"
	POST   httpMethod = "POST"
	PUT    httpMethod = "PUT"
	DELETE httpMethod = "DELETE"
)

func (h httpMethod) ToString() string {
	return string(h)
}

type httpHandlerArgs struct {
	URL         string
	Method      httpMethod
	Payload     any
	Credentials Credentials
}

func httpHandler(ctx context.Context, args httpHandlerArgs) ([]byte, error) {
	// Request URL
	url := fmt.Sprintf("%s%s", args.Credentials.BaseURL, args.URL)

	// Request body
	var body io.Reader
	if args.Payload != nil {
		payloadBytes, err := json.Marshal(args.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to marshal the request payload")
		}
		body = bytes.NewReader(payloadBytes)
	}

	// Request
	req, err := http.NewRequestWithContext(ctx, args.Method.ToString(), url, body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create new http request")
	}

	// Request Headers
	req.Header.Add("Content-Type", "application/json")

	if args.Credentials.AccessToken == "" {
		return nil, errors.Wrap(err, "Access token is required")
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", args.Credentials.AccessToken))

	// Make request
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, errors.Wrap(err, "Request canceled or timed out")
		}
		return nil, errors.Wrap(err, "Failed to make the HTTP request")
	}
	defer res.Body.Close()

	responseBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read the response body")
	}

	// Check for error
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status code %d: %s", res.StatusCode, string(responseBytes))
	}

	return responseBytes, nil
}

type Customer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) EnableStardustAI(ctx context.Context) (Customer, error) {
	url := "/api/v1/user/enable-stardust-ai"

	args := httpHandlerArgs{
		URL:         url,
		Method:      POST,
		Credentials: c.Credentials,
	}

	res, err := httpHandler(ctx, args)
	if err != nil {
		return Customer{}, errors.Wrap(err, "Failed to enable Stardust AI")
	}

	var enableStardustAIRes struct {
		Success    bool     `json:"success"`
		Message    string   `json:"message"`
		StatusCode int      `json:"status_code"`
		Data       Customer `json:"data"`
	}
	err = json.Unmarshal(res, &enableStardustAIRes)
	if err != nil {
		return Customer{}, errors.Wrap(err, "Failed to unmarshal customer response")
	}

	return enableStardustAIRes.Data, nil
}

func (c *Client) DisableStardustAI(ctx context.Context) (bool, error) {
	url := "/api/v1/user/disable-stardust-ai"

	args := httpHandlerArgs{
		URL:         url,
		Method:      POST,
		Credentials: c.Credentials,
	}

	res, err := httpHandler(ctx, args)
	if err != nil {
		return false, errors.Wrap(err, "Failed to disable Stardust AI")
	}

	var disableStardustAIRes struct {
		Success    bool   `json:"success"`
		Message    string `json:"message"`
		StatusCode int    `json:"status_code"`
		Data       bool   `json:"data"`
	}
	err = json.Unmarshal(res, &disableStardustAIRes)
	if err != nil {
		return false, errors.Wrap(err, "Failed to unmarshal response")
	}

	return disableStardustAIRes.Data, nil
}

func (c *Client) SwitchDefaultKey(ctx context.Context, switchValue bool) (bool, error) {
	switchValueStr := "false"
	if switchValue {
		switchValueStr = "true"
	}
	url := fmt.Sprintf("/api/v1/user/switch-default-key?switch=%s", switchValueStr)

	args := httpHandlerArgs{
		URL:         url,
		Method:      POST,
		Credentials: c.Credentials,
	}

	res, err := httpHandler(ctx, args)
	if err != nil {
		return false, errors.Wrap(err, "Failed to enable Stardust AI")
	}

	var switchDefaultKeyRes struct {
		Success    bool   `json:"success"`
		Message    string `json:"message"`
		StatusCode int    `json:"status_code"`
		Data       bool   `json:"data"`
	}
	err = json.Unmarshal(res, &switchDefaultKeyRes)
	if err != nil {
		return false, errors.Wrap(err, "Failed to unmarshal customer response")
	}

	return switchDefaultKeyRes.Data, nil
}

type UserProvider struct {
	KeyID    string `json:"key_id"`
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
}

func (c *Client) ListUserProviders(ctx context.Context) ([]UserProvider, error) {
	url := "/api/v1/user/providers/list"

	args := httpHandlerArgs{
		URL:         url,
		Method:      GET,
		Credentials: c.Credentials,
	}

	res, err := httpHandler(ctx, args)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch user providers")
	}

	var ListUserProvidersRes struct {
		Success    bool           `json:"success"`
		Message    string         `json:"message"`
		StatusCode int            `json:"status_code"`
		Data       []UserProvider `json:"data"`
	}
	err = json.Unmarshal(res, &ListUserProvidersRes)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal user providers")
	}

	return ListUserProvidersRes.Data, nil
}

type AddProviderAPIKeyReq struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
}

type VirtualKey struct {
	ID              string           `json:"id"`
	Value           string           `json:"value"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	CustomerID      string           `json:"customer_id"`
	IsActive        bool             `json:"is_active"`
	ProviderConfigs []ProviderConfig `json:"provider_configs"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProviderConfig struct {
	ID                int64    `json:"id"`
	VirtualKeyID      string   `json:"virtual_key_id"`
	Provider          string   `json:"provider"`
	Weight            *int64   `json:"weight"`
	AllowedModels     []string `json:"allowed_models"`
	BlacklistedModels []string `json:"blacklisted_models"`
	AllowAllKeys      bool     `json:"allow_all_keys"`
	Keys              []Key    `json:"keys"`
}

type Key struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Value      Value     `json:"value"`
	ProviderID int64     `json:"provider_id"`
	Provider   string    `json:"provider"`
	KeyID      string    `json:"key_id"`
	Models     []string  `json:"models"`
	Weight     int64     `json:"weight"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Value struct {
	Value   string `json:"value"`
	EnvVar  string `json:"env_var"`
	FromEnv bool   `json:"from_env"`
}

func (c *Client) AddProviderAPIKey(ctx context.Context, r AddProviderAPIKeyReq) (VirtualKey, error) {
	url := "/api/v1/providers/api-key"

	args := httpHandlerArgs{
		URL:         url,
		Method:      POST,
		Payload:     r,
		Credentials: c.Credentials,
	}

	res, err := httpHandler(ctx, args)
	if err != nil {
		return VirtualKey{}, errors.Wrap(err, "Failed to add provider api key")
	}

	var AddProviderAPIKeyRes struct {
		Success    bool       `json:"success"`
		Message    string     `json:"message"`
		StatusCode int        `json:"status_code"`
		Data       VirtualKey `json:"data"`
	}
	err = json.Unmarshal(res, &AddProviderAPIKeyRes)
	if err != nil {
		return VirtualKey{}, errors.Wrap(err, "Failed to unmarshal virtual key")
	}

	return AddProviderAPIKeyRes.Data, nil
}

type UpdateProviderAPIKeyReq struct {
	KeyID    string `json:"key_id"`
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
}

func (c *Client) UpdateProviderAPIKey(ctx context.Context, r UpdateProviderAPIKeyReq) (bool, error) {
	url := "/api/v1/providers/api-key"

	args := httpHandlerArgs{
		URL:         url,
		Method:      PUT,
		Payload:     r,
		Credentials: c.Credentials,
	}

	res, err := httpHandler(ctx, args)
	if err != nil {
		return false, errors.Wrap(err, "Failed to update provider api key")
	}

	var UpdateProviderAPIKeyRes struct {
		Success    bool   `json:"success"`
		Message    string `json:"message"`
		StatusCode int    `json:"status_code"`
		Data       bool   `json:"data"`
	}
	err = json.Unmarshal(res, &UpdateProviderAPIKeyRes)
	if err != nil {
		return false, errors.Wrap(err, "Failed to unmarshal provider key")
	}

	return UpdateProviderAPIKeyRes.Data, nil
}

type DeleteProviderAPIKeyReq struct {
	KeyID    string `json:"key_id"`
	Provider string `json:"provider"`
}

func (c *Client) DeleteProviderAPIKey(ctx context.Context, r DeleteProviderAPIKeyReq) (bool, error) {
	url := fmt.Sprintf("/api/v1/providers/api-key?provider=%s&key_id=%s", r.Provider, r.KeyID)

	args := httpHandlerArgs{
		URL:         url,
		Method:      DELETE,
		Payload:     r,
		Credentials: c.Credentials,
	}

	res, err := httpHandler(ctx, args)
	if err != nil {
		return false, errors.Wrap(err, "Failed to delete provider api key")
	}

	var DeleteProviderAPIKeyRes struct {
		Success    bool   `json:"success"`
		Message    string `json:"message"`
		StatusCode int    `json:"status_code"`
		Data       bool   `json:"data"`
	}
	err = json.Unmarshal(res, &DeleteProviderAPIKeyRes)
	if err != nil {
		return false, errors.Wrap(err, "Failed to unmarshal provider key")
	}

	return DeleteProviderAPIKeyRes.Data, nil
}
