package api

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// ─── ListSubscriptions ────────────────────────────────────────────────────────

func TestListSubscriptions_HappyPath(t *testing.T) {
	const want = `{"subscriptions":[{"id":1}]}`
	c := newTestClient(t, respond(t, http.MethodGet, "/subscriptions/", 200, want))
	got, err := c.ListSubscriptions(ListSubscriptionsParams{})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("body: got %q, want %q", got, want)
	}
}

func TestListSubscriptions_QueryParams(t *testing.T) {
	var capturedQuery url.Values
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.Query()
		w.WriteHeader(200)
		_, _ = io.WriteString(w, `{}`)
	})
	_, _ = c.ListSubscriptions(ListSubscriptionsParams{
		Query:               "alice",
		Page:                2,
		ActiveOnly:          true,
		PausedOnly:          false,
		CancelledOnly:       false,
		WithItems:           true,
		WithBillingAttempts: true,
	})
	if got := capturedQuery.Get("query"); got != "alice" {
		t.Errorf("query param: got %q, want %q", got, "alice")
	}
	if got := capturedQuery.Get("page"); got != "2" {
		t.Errorf("page param: got %q, want %q", got, "2")
	}
	if got := capturedQuery.Get("active-only"); got != "true" {
		t.Errorf("active-only param: got %q, want true", got)
	}
	if got := capturedQuery.Get("with-items"); got != "true" {
		t.Errorf("with-items param: got %q, want true", got)
	}
	if got := capturedQuery.Get("with-billing-attempts"); got != "true" {
		t.Errorf("with-billing-attempts param: got %q, want true", got)
	}
	// paused-only and cancelled-only should NOT be set
	if got := capturedQuery.Get("paused-only"); got != "" {
		t.Errorf("paused-only should be absent, got %q", got)
	}
}

func TestListSubscriptions_Error(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
		_, _ = io.WriteString(w, `internal error`)
	})
	_, err := c.ListSubscriptions(ListSubscriptionsParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ─── GetSubscription ──────────────────────────────────────────────────────────

func TestGetSubscription_HappyPath(t *testing.T) {
	const want = `{"id":42}`
	var capturedID string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedID = r.URL.Query().Get("id")
		w.WriteHeader(200)
		_, _ = io.WriteString(w, want)
	})
	got, err := c.GetSubscription(42)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("body: got %q, want %q", got, want)
	}
	if capturedID != "42" {
		t.Errorf("id query param: got %q, want 42", capturedID)
	}
}

func TestGetSubscription_NotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		_, _ = io.WriteString(w, `{"error":"not found"}`)
	})
	_, err := c.GetSubscription(999)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 error, got: %v", err)
	}
}

// ─── CancelSubscription ───────────────────────────────────────────────────────

func TestCancelSubscription_HappyPath(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			assertBodyField(t, body, "action", "cancel")
			assertBodyFieldFloat(t, body, "id", 10)
		},
	))
	got, err := c.CancelSubscription(10)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"success":true}` {
		t.Errorf("unexpected body: %s", got)
	}
}

// ─── PauseSubscription ────────────────────────────────────────────────────────

func TestPauseSubscription_HappyPath(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			assertBodyField(t, body, "action", "pause")
			assertBodyFieldFloat(t, body, "id", 20)
		},
	))
	_, err := c.PauseSubscription(20)
	if err != nil {
		t.Fatal(err)
	}
}

// ─── ReactivateSubscription ───────────────────────────────────────────────────

func TestReactivateSubscription_HappyPath(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			assertBodyField(t, body, "action", "reactivate")
		},
	))
	_, err := c.ReactivateSubscription(30)
	if err != nil {
		t.Fatal(err)
	}
}

// ─── ResumeSubscription ───────────────────────────────────────────────────────

func TestResumeSubscription_HappyPath(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			assertBodyField(t, body, "action", "resume")
		},
	))
	_, err := c.ResumeSubscription(40)
	if err != nil {
		t.Fatal(err)
	}
}

// ─── EditSubscription ─────────────────────────────────────────────────────────

func TestEditSubscription_HappyPath(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			assertBodyField(t, body, "action", "edit")
			edit, ok := body["edit"].(map[string]any)
			if !ok {
				t.Error("missing 'edit' field in body")
				return
			}
			if edit["delivery_interval"] != "2 week" {
				t.Errorf("delivery_interval: got %v, want '2 week'", edit["delivery_interval"])
			}
		},
	))
	_, err := c.EditSubscription(50, map[string]interface{}{
		"delivery_interval": "2 week",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEditSubscription_ApiError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(422)
		_, _ = io.WriteString(w, `{"error":"invalid interval"}`)
	})
	_, err := c.EditSubscription(50, map[string]interface{}{"delivery_interval": "bad"})
	if err == nil || !strings.Contains(err.Error(), "422") {
		t.Errorf("expected 422 error, got: %v", err)
	}
}

// ─── AddItems ─────────────────────────────────────────────────────────────────

func TestAddItems_HappyPath(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			assertBodyField(t, body, "action", "add_items")
			assertBodyFieldFloat(t, body, "id", 60)
			items, ok := body["add_items"].([]any)
			if !ok || len(items) != 1 {
				t.Errorf("add_items: expected 1 item, got: %v", body["add_items"])
				return
			}
			item := items[0].(map[string]any)
			if item["title"] != "Coffee 1kg" {
				t.Errorf("item title: got %v, want Coffee 1kg", item["title"])
			}
		},
	))
	_, err := c.AddItems(60, []SubscriptionItem{
		{ProductID: "1", VariantID: "2", Title: "Coffee 1kg", Price: 24.00, Quantity: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ─── RemoveItems ──────────────────────────────────────────────────────────────

func TestRemoveItems_HappyPath(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			assertBodyField(t, body, "action", "remove_items")
			ids, ok := body["remove_items"].([]any)
			if !ok || len(ids) != 2 {
				t.Errorf("remove_items: expected 2 IDs, got: %v", body["remove_items"])
			}
		},
	))
	_, err := c.RemoveItems(70, []int{101, 102})
	if err != nil {
		t.Fatal(err)
	}
}

// ─── Assertion helpers ────────────────────────────────────────────────────────

func assertBodyField(t *testing.T, body map[string]any, key, want string) {
	t.Helper()
	got, ok := body[key].(string)
	if !ok {
		t.Errorf("body[%q]: not a string, got %T (%v)", key, body[key], body[key])
		return
	}
	if got != want {
		t.Errorf("body[%q]: got %q, want %q", key, got, want)
	}
}

func assertBodyFieldFloat(t *testing.T, body map[string]any, key string, want float64) {
	t.Helper()
	got, ok := body[key].(float64)
	if !ok {
		t.Errorf("body[%q]: not a number, got %T (%v)", key, body[key], body[key])
		return
	}
	if got != want {
		t.Errorf("body[%q]: got %v, want %v", key, got, want)
	}
}
