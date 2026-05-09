package output

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// statusWeight returns a sort priority for failure-first ordering.
// Lower = shown first. Non-active states surface before ACTIVE.
func statusWeight(s string) int {
	switch s {
	case "FAILED", "PAYMENT_FAILED":
		return 0
	case "PAUSED":
		return 1
	case "CANCELLED":
		return 2
	default: // ACTIVE and anything else
		return 3
	}
}

// AgentSubscriptionDetail renders a single subscription as one compact JSON
// line with only whitelisted fields.
func AgentSubscriptionDetail(data []byte) error {
	var resp struct {
		Payload struct {
			ID            int64  `json:"id"`
			Status        string `json:"status"`
			Email         string `json:"email"`
			FirstName     string `json:"first_name"`
			LastName      string `json:"last_name"`
			Currency      string `json:"currency"`
			Total         float64 `json:"total_value"`
			Interval      string `json:"delivery_interval"`
			SAddress      string `json:"s_address1"`
			SCity         string `json:"s_city"`
			SCountry      string `json:"s_country"`
			BAddress      string `json:"b_address1"`
			BCity         string `json:"b_city"`
			BCountry      string `json:"b_country"`
			CardBrand     string `json:"card_brand"`
			CardLast      string `json:"card_last_digits"`
			CardExpM      string `json:"card_expiry_month"`
			CardExpY      string `json:"card_expiry_year"`
			Items []struct {
				ID    int64  `json:"id"`
				Title string `json:"title"`
				Qty   int    `json:"quantity"`
				Price string `json:"price"`
			} `json:"items"`
			Attempts []struct {
				ID     int64  `json:"id"`
				Date   string `json:"date"`
				Status string `json:"status"`
			} `json:"billing_attempts"`
			Invoices []struct {
				ID            int64  `json:"id"`
				Date          string `json:"date"`
				PaymentStatus string `json:"payment_status"`
			} `json:"invoices"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return printRaw(data)
	}
	p := resp.Payload

	out := agentSubscription{
		ID:            p.ID,
		Status:        p.Status,
		CustomerEmail: p.Email,
		CustomerName:  strings.TrimSpace(p.FirstName + " " + strings.TrimSpace(p.LastName)),
		Currency:      p.Currency,
		Total:         fmt.Sprintf("%.2f", p.Total),
		Interval:      p.Interval,
	}

	// Address: shipping preferred, fall back to billing
	addr := p.SAddress
	city := p.SCity
	country := p.SCountry
	if addr == "" {
		addr = p.BAddress
		city = p.BCity
		country = p.BCountry
	}
	if addr != "" {
		out.Address = addr + ", " + city + ", " + country
	}

	// Card summary
	if p.CardBrand != "" {
		out.CardSummary = fmt.Sprintf("%s *%s exp %s/%s", p.CardBrand, p.CardLast, p.CardExpM, p.CardExpY)
	}

	// Items
	for _, it := range p.Items {
		out.Items = append(out.Items, agentItem{
			ID:    it.ID,
			Title: it.Title,
			Qty:   it.Qty,
			Price: it.Price,
		})
	}

	// Next billing attempts (auto_charge): skip completed past attempts
	for _, a := range p.Attempts {
		if a.Status == "completed" {
			continue
		}
		status := a.Status
		if status == "" {
			status = "scheduled"
		}
		out.NextAttempts = append(out.NextAttempts, agentAttempt{
			ID:     a.ID,
			Date:   a.Date,
			Status: status,
		})
	}

	// Invoices (recurring_invoice payment type)
	for _, inv := range p.Invoices {
		out.Invoices = append(out.Invoices, agentInvoice{
			ID:            inv.ID,
			Date:          inv.Date,
			PaymentStatus: inv.PaymentStatus,
		})
	}

	return writeCompactJSON(out)
}

// AgentSubscriptionList renders subscriptions as NDJSON (one compact JSON
// object per line) with non-active statuses surfaced first, followed by a
// _meta line.
func AgentSubscriptionList(data []byte) error {
	var resp struct {
		Payload struct {
			Page        int `json:"page"`
			TotalPages  int `json:"total_pages"`
			Subscriptions []struct {
				ID       int64   `json:"id"`
				Status   string  `json:"status"`
				Email    string  `json:"email"`
				Currency string  `json:"currency"`
				Total    float64 `json:"total_value"`
				Interval string  `json:"delivery_interval"`
			} `json:"subscriptions"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return printRaw(data)
	}
	p := resp.Payload

	// Failure-first sort (stable: preserves original order within same weight)
	sort.SliceStable(p.Subscriptions, func(i, j int) bool {
		return statusWeight(p.Subscriptions[i].Status) < statusWeight(p.Subscriptions[j].Status)
	})

	for _, s := range p.Subscriptions {
		row := agentListItem{
			ID:            s.ID,
			Status:        s.Status,
			CustomerEmail: s.Email,
			Currency:      s.Currency,
			Total:         fmt.Sprintf("%.2f", s.Total),
			Interval:      s.Interval,
		}
		if err := writeCompactJSON(row); err != nil {
			return err
		}
	}

	// _meta line
	meta := agentMeta{}
	meta.Meta.Page = p.Page
	meta.Meta.Count = len(p.Subscriptions)
	meta.Meta.HasMore = p.TotalPages > p.Page
	return writeCompactJSON(meta)
}

// AgentSuccess writes a minimal mutation acknowledgement: {"ok":true,"id":N,"action":"verb"}.
func AgentSuccess(data []byte, action string, id int) error {
	ack := agentSuccessAck{Ok: true, ID: id, Action: action}
	return writeCompactJSON(ack)
}

// AgentError writes the full API error body to a temp file, then returns a
// compact JSON error envelope (as []byte) that callers can print.
// It does NOT write to stdout — the caller decides where to send it.
func AgentError(body []byte, statusCode int) []byte {
	// Extract error message if possible
	var apiErr struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(body, &apiErr)

	// Write full body to temp file
	tmpPath := ""
	name := fmt.Sprintf("seal-cli-err-%04x.json", rand.Intn(0xFFFF))
	path := filepath.Join(os.TempDir(), name)
	if err := os.WriteFile(path, body, 0600); err == nil {
		tmpPath = path
	}

	ack := agentErrorAck{
		Ok:     false,
		Status: statusCode,
		Error:  apiErr.Error,
		Detail: tmpPath,
	}
	b, _ := json.Marshal(ack)
	return b
}

// writeCompactJSON marshals v as compact JSON and writes it as one line to stdout.
func writeCompactJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stdout, "%s\n", b)
	return err
}
