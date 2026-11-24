package dbmx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// Billing is the user's plan and subscription state.
type Billing struct {
	Plan                    string     `json:"plan"`
	Status                  string     `json:"status"`
	NextBillingDate         *time.Time `json:"next_billing_date"`
	CancelAtNextBillingDate bool       `json:"cancel_at_next_billing_date"`
	TrialEndsAt             *time.Time `json:"trial_ends_at"`
	TrialAvailable          bool       `json:"trial_available"`
	AIEntitled              bool       `json:"ai_entitled"`
	TeamCoversAI            bool       `json:"team_covers_ai"`
}

func (c *Client) billingStatus(ctx context.Context, method httpMethod, url string, payload any) (Billing, error) {
	res, err := httpHandler(ctx, httpHandlerArgs{URL: url, Method: method, Payload: payload, Credentials: c.Credentials})
	if err != nil {
		return Billing{}, err
	}

	var out struct {
		Success    bool    `json:"success"`
		Message    string  `json:"message"`
		StatusCode int     `json:"status_code"`
		Data       Billing `json:"data"`
	}
	if err := json.Unmarshal(res, &out); err != nil {
		return Billing{}, errors.Wrap(err, "Failed to unmarshal billing response")
	}
	return out.Data, nil
}

func (c *Client) billingURL(ctx context.Context, url string, payload any) (string, error) {
	res, err := httpHandler(ctx, httpHandlerArgs{URL: url, Method: POST, Payload: payload, Credentials: c.Credentials})
	if err != nil {
		return "", err
	}

	var out struct {
		Success    bool   `json:"success"`
		Message    string `json:"message"`
		StatusCode int    `json:"status_code"`
		Data       struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(res, &out); err != nil {
		return "", errors.Wrap(err, "Failed to unmarshal billing response")
	}
	return out.Data.URL, nil
}

type planReq struct {
	Plan string `json:"plan"`
}

func (c *Client) GetBilling(ctx context.Context) (Billing, error) {
	b, err := c.billingStatus(ctx, GET, "/api/v1/billing", nil)
	return b, errors.Wrap(err, "Failed to get billing")
}

func (c *Client) StartTrial(ctx context.Context, plan string) (Billing, error) {
	b, err := c.billingStatus(ctx, POST, "/api/v1/billing/trial", planReq{Plan: plan})
	return b, errors.Wrap(err, "Failed to start trial")
}

func (c *Client) CreateCheckout(ctx context.Context, plan string) (string, error) {
	u, err := c.billingURL(ctx, "/api/v1/billing/checkout", planReq{Plan: plan})
	return u, errors.Wrap(err, "Failed to create checkout")
}

func (c *Client) CreatePortal(ctx context.Context) (string, error) {
	u, err := c.billingURL(ctx, "/api/v1/billing/portal", nil)
	return u, errors.Wrap(err, "Failed to create portal")
}
