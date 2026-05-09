package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://app.sealsubscriptions.com/shopify/merchant/api"

// Client wraps the Seal Subscriptions API.
type Client struct {
	token      string
	httpClient *http.Client
}

// New creates a new API client with the given token.
func New(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout:       30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error { return nil },
		},
	}
}

func (c *Client) do(method, endpoint string, query url.Values, body interface{}) ([]byte, error) {
	u := baseURL + endpoint + "/"
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Seal-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode == 503 {
		return nil, fmt.Errorf("API rate limit exceeded (503). Retry after current requests finish")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func (c *Client) get(endpoint string, query url.Values) ([]byte, error) {
	return c.do(http.MethodGet, endpoint, query, nil)
}

func (c *Client) post(endpoint string, body interface{}) ([]byte, error) {
	return c.do(http.MethodPost, endpoint, nil, body)
}

func (c *Client) put(endpoint string, body interface{}) ([]byte, error) {
	return c.do(http.MethodPut, endpoint, nil, body)
}

func (c *Client) delete(endpoint string, query url.Values) ([]byte, error) {
	return c.do(http.MethodDelete, endpoint, query, nil)
}

// ─── Subscriptions ──────────────────────────────────────────────────────────

type ListSubscriptionsParams struct {
	Query             string
	Page              int
	ActiveOnly        bool
	PausedOnly        bool
	CancelledOnly     bool
	WithItems         bool
	WithBillingAttempts bool
}

func (c *Client) ListSubscriptions(p ListSubscriptionsParams) ([]byte, error) {
	q := url.Values{}
	if p.Query != "" {
		q.Set("query", p.Query)
	}
	if p.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", p.Page))
	}
	if p.ActiveOnly {
		q.Set("active-only", "true")
	}
	if p.PausedOnly {
		q.Set("paused-only", "true")
	}
	if p.CancelledOnly {
		q.Set("cancelled-only", "true")
	}
	if p.WithItems {
		q.Set("with-items", "true")
	}
	if p.WithBillingAttempts {
		q.Set("with-billing-attempts", "true")
	}
	return c.get("/subscriptions", q)
}

func (c *Client) GetSubscription(id int) ([]byte, error) {
	q := url.Values{}
	q.Set("id", fmt.Sprintf("%d", id))
	return c.get("/subscription", q)
}

func (c *Client) CancelSubscription(id int) ([]byte, error) {
	return c.put("/subscription", map[string]interface{}{
		"id":     id,
		"action": "cancel",
	})
}

func (c *Client) PauseSubscription(id int) ([]byte, error) {
	return c.put("/subscription", map[string]interface{}{
		"id":     id,
		"action": "pause",
	})
}

func (c *Client) ReactivateSubscription(id int) ([]byte, error) {
	return c.put("/subscription", map[string]interface{}{
		"id":     id,
		"action": "reactivate",
	})
}

func (c *Client) ResumeSubscription(id int) ([]byte, error) {
	return c.put("/subscription", map[string]interface{}{
		"id":     id,
		"action": "resume",
	})
}

func (c *Client) EditSubscription(id int, edit map[string]interface{}) ([]byte, error) {
	return c.put("/subscription", map[string]interface{}{
		"id":     id,
		"action": "edit",
		"edit":   edit,
	})
}

// ─── Billing Attempts ────────────────────────────────────────────────────────

func (c *Client) RescheduleBillingAttempt(id, subscriptionID int, date, timeStr, timezone string, resetSchedule bool) ([]byte, error) {
	body := map[string]interface{}{
		"id":              id,
		"subscription_id": subscriptionID,
		"date":            date,
		"time":            timeStr,
		"timezone":        timezone,
		"action":          "reschedule",
	}
	if resetSchedule {
		body["reset_schedule"] = "true"
	}
	return c.put("/subscription-billing-attempt", body)
}

func (c *Client) DeleteBillingAttempt(id, subscriptionID int) ([]byte, error) {
	q := url.Values{}
	q.Set("id", fmt.Sprintf("%d", id))
	q.Set("subscription_id", fmt.Sprintf("%d", subscriptionID))
	return c.delete("/subscription-billing-attempt", q)
}

func (c *Client) SkipBillingAttempt(id, subscriptionID int) ([]byte, error) {
	return c.put("/subscription-billing-attempt", map[string]interface{}{
		"id":              id,
		"subscription_id": subscriptionID,
		"action":          "skip",
	})
}

func (c *Client) UnskipBillingAttempt(id, subscriptionID int) ([]byte, error) {
	return c.put("/subscription-billing-attempt", map[string]interface{}{
		"id":              id,
		"subscription_id": subscriptionID,
		"action":          "unskip",
	})
}
