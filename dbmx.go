package dbmx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

type Response struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Data       any    `json:"data"`
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

	var response Response
	err = json.Unmarshal(res, &response)
	if err != nil {
		return Customer{}, errors.Wrap(err, "Failed to unmarshal customer response")
	}

	customer := response.Data.(Customer)

	return customer, nil
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

	var response Response
	err = json.Unmarshal(res, &response)
	if err != nil {
		return false, errors.Wrap(err, "Failed to unmarshal response")
	}

	return response.Data.(bool), nil
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

	var response Response
	err = json.Unmarshal(res, &response)
	if err != nil {
		return false, errors.Wrap(err, "Failed to unmarshal customer response")
	}

	return response.Data.(bool), nil
}
