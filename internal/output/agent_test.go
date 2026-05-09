package output

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// captureStdout redirects os.Stdout for the duration of fn and returns what
// was written to it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	fn()

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	r.Close()
	return buf.String()
}

// ─── AgentSubscriptionDetail ──────────────────────────────────────────────────

// Minimal realistic API response for a single subscription GET.
const agentSubDetailFixture = `{
	"success": true,
	"payload": {
		"id": 12345,
		"status": "ACTIVE",
		"email": "jane@example.com",
		"first_name": "Jane",
		"last_name": "Doe",
		"currency": "USD",
		"total_value": 29.00,
		"delivery_interval": "1 month",
		"billing_interval": "1 month",
		"s_address1": "123 Main St",
		"s_city": "Austin",
		"s_country": "United States",
		"b_address1": "",
		"b_city": "",
		"b_country": "",
		"card_brand": "visa",
		"card_last_digits": "4242",
		"card_expiry_month": "12",
		"card_expiry_year": "2027",
		"items": [
			{"id": 111, "title": "Coffee Box", "quantity": 2, "price": "14.50", "final_price": "14.50"}
		],
		"billing_attempts": [
			{"id": 999, "date": "2026-06-01T08:00:00+00:00", "status": ""}
		]
	}
}`

func TestAgentSubscriptionDetail_IsCompactJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := AgentSubscriptionDetail([]byte(agentSubDetailFixture)); err != nil {
			t.Errorf("AgentSubscriptionDetail returned error: %v", err)
		}
	})

	// Must be a single non-empty line terminated with newline
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d:\n%s", len(lines), out)
	}

	// Must be valid JSON
	var m map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %q", err, out)
	}

	// Must contain essential whitelisted fields
	for _, key := range []string{"id", "status", "customer_email", "currency", "total", "interval"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing field %q in agent output", key)
		}
	}

	// Must NOT contain internal/noisy fields
	for _, key := range []string{"edit_url", "shopify_graphql_subscription_contract_id",
		"payment_method_id", "note_attributes", "log", "fulfillment_orders"} {
		if _, ok := m[key]; ok {
			t.Errorf("field %q should be excluded from agent output", key)
		}
	}

	// id should be 12345
	if got, _ := m["id"].(float64); got != 12345 {
		t.Errorf("id: got %v, want 12345", got)
	}
}

func TestAgentSubscriptionDetail_ItemsPresent(t *testing.T) {
	out := captureStdout(t, func() {
		_ = AgentSubscriptionDetail([]byte(agentSubDetailFixture))
	})
	var m map[string]any
	_ = json.Unmarshal([]byte(strings.TrimSpace(out)), &m)

	items, ok := m["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected items array with entries, got: %v", m["items"])
	}
	item := items[0].(map[string]any)
	if item["title"] != "Coffee Box" {
		t.Errorf("item title: got %v, want 'Coffee Box'", item["title"])
	}
}

func TestAgentSubscriptionDetail_NoIndentation(t *testing.T) {
	out := captureStdout(t, func() {
		_ = AgentSubscriptionDetail([]byte(agentSubDetailFixture))
	})
	// Compact JSON must not contain "  " (2-space indent) or newlines inside
	if strings.Contains(strings.TrimRight(out, "\n"), "\n") {
		t.Error("output contains embedded newlines — not compact JSON")
	}
	if strings.Contains(out, "  ") {
		t.Error("output contains indentation — not compact JSON")
	}
}

// ─── AgentSubscriptionList ────────────────────────────────────────────────────

const agentSubListFixture = `{
	"success": true,
	"payload": {
		"page": 1,
		"total_pages": 2,
		"subscriptions": [
			{"id": 1, "status": "ACTIVE",    "email": "a@x.com", "first_name": "A", "last_name": "X", "currency": "USD", "total_value": 10, "delivery_interval": "1 month"},
			{"id": 2, "status": "PAUSED",    "email": "b@x.com", "first_name": "B", "last_name": "X", "currency": "USD", "total_value": 20, "delivery_interval": "1 month"},
			{"id": 3, "status": "CANCELLED", "email": "c@x.com", "first_name": "C", "last_name": "X", "currency": "USD", "total_value": 30, "delivery_interval": "1 month"}
		]
	}
}`

func TestAgentSubscriptionList_IsNDJSON(t *testing.T) {
	out := captureStdout(t, func() {
		if err := AgentSubscriptionList([]byte(agentSubListFixture)); err != nil {
			t.Errorf("AgentSubscriptionList returned error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// 3 subscription lines + 1 _meta line
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines (3 subs + _meta), got %d:\n%s", len(lines), out)
	}

	// Every line must be valid JSON
	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Errorf("line %d is not valid JSON: %v\n  line: %q", i, err, line)
		}
	}
}

func TestAgentSubscriptionList_FailureFirstOrder(t *testing.T) {
	out := captureStdout(t, func() {
		_ = AgentSubscriptionList([]byte(agentSubListFixture))
	})

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// First non-meta line should be PAUSED (non-active) before ACTIVE
	// Order expected: PAUSED(2), CANCELLED(3), ACTIVE(1)
	firstSub := map[string]any{}
	_ = json.Unmarshal([]byte(lines[0]), &firstSub)
	status, _ := firstSub["status"].(string)
	if status == "ACTIVE" {
		t.Errorf("failure-first ordering violated: first record has status ACTIVE, expected PAUSED or CANCELLED")
	}
}

func TestAgentSubscriptionList_MetaLine(t *testing.T) {
	out := captureStdout(t, func() {
		_ = AgentSubscriptionList([]byte(agentSubListFixture))
	})

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	lastLine := lines[len(lines)-1]

	var m map[string]any
	_ = json.Unmarshal([]byte(lastLine), &m)
	meta, ok := m["_meta"].(map[string]any)
	if !ok {
		t.Fatalf("last line missing '_meta' key: %s", lastLine)
	}
	if meta["page"] != float64(1) {
		t.Errorf("_meta.page: got %v, want 1", meta["page"])
	}
	if meta["has_more"] != true {
		t.Errorf("_meta.has_more: got %v, want true (total_pages=2)", meta["has_more"])
	}
}

// ─── AgentSuccess ─────────────────────────────────────────────────────────────

func TestAgentSuccess_OkTrue(t *testing.T) {
	out := captureStdout(t, func() {
		if err := AgentSuccess([]byte(`{"success":true,"message":"Cancelled"}`), "cancel", 12345); err != nil {
			t.Errorf("AgentSuccess returned error: %v", err)
		}
	})

	var m map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("AgentSuccess output not valid JSON: %v\noutput: %q", err, out)
	}
	if m["ok"] != true {
		t.Errorf("ok: got %v, want true", m["ok"])
	}
	if m["id"] != float64(12345) {
		t.Errorf("id: got %v, want 12345", m["id"])
	}
	if m["action"] != "cancel" {
		t.Errorf("action: got %v, want 'cancel'", m["action"])
	}
}

// ─── AgentError ───────────────────────────────────────────────────────────────

func TestAgentError_WritesToTempAndReturnsPath(t *testing.T) {
	body := []byte(`{"success":false,"error":"Subscription was not found."}`)
	result := AgentError(body, 404)

	var m map[string]any
	if err := json.Unmarshal(result, &m); err != nil {
		t.Fatalf("AgentError output not valid JSON: %v\noutput: %q", err, result)
	}
	if m["ok"] != false {
		t.Errorf("ok: got %v, want false", m["ok"])
	}
	if m["status"] != float64(404) {
		t.Errorf("status: got %v, want 404", m["status"])
	}
	detail, ok := m["detail"].(string)
	if !ok || detail == "" {
		t.Errorf("detail: expected a non-empty file path, got %v", m["detail"])
	}
	// The temp file should actually exist and contain the original body
	if _, err := os.Stat(detail); err != nil {
		t.Errorf("temp file %q does not exist: %v", detail, err)
	}
	t.Cleanup(func() { os.Remove(detail) })
}
