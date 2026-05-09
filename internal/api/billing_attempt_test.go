package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// ─── RescheduleBillingAttempt ─────────────────────────────────────────────────

func TestRescheduleBillingAttempt_HappyPath(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription-billing-attempt/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			assertBodyField(t, body, "action", "reschedule")
			assertBodyFieldFloat(t, body, "id", 1)
			assertBodyFieldFloat(t, body, "subscription_id", 100)
			assertBodyField(t, body, "date", "2025-12-01")
			assertBodyField(t, body, "time", "14:30")
			assertBodyField(t, body, "timezone", "+00:00")
		},
	))
	_, err := c.RescheduleBillingAttempt(1, 100, "2025-12-01", "14:30", "+00:00", false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRescheduleBillingAttempt_WithResetSchedule(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription-billing-attempt/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			if body["reset_schedule"] != "true" {
				t.Errorf("reset_schedule: got %v, want 'true'", body["reset_schedule"])
			}
		},
	))
	_, err := c.RescheduleBillingAttempt(1, 100, "2025-12-01", "14:30", "+00:00", true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRescheduleBillingAttempt_ApiError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(422)
		_, _ = io.WriteString(w, `{"error":"invalid date"}`)
	})
	_, err := c.RescheduleBillingAttempt(1, 100, "bad-date", "14:30", "+00:00", false)
	if err == nil || !strings.Contains(err.Error(), "422") {
		t.Errorf("expected 422 error, got: %v", err)
	}
}

// ─── DeleteBillingAttempt ─────────────────────────────────────────────────────

func TestDeleteBillingAttempt_HappyPath(t *testing.T) {
	var capturedQuery map[string]string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method: got %q, want DELETE", r.Method)
		}
		capturedQuery = map[string]string{
			"id":              r.URL.Query().Get("id"),
			"subscription_id": r.URL.Query().Get("subscription_id"),
		}
		w.WriteHeader(200)
		_, _ = io.WriteString(w, `{"success":true}`)
	})
	_, err := c.DeleteBillingAttempt(5, 200)
	if err != nil {
		t.Fatal(err)
	}
	if capturedQuery["id"] != "5" {
		t.Errorf("id param: got %q, want 5", capturedQuery["id"])
	}
	if capturedQuery["subscription_id"] != "200" {
		t.Errorf("subscription_id param: got %q, want 200", capturedQuery["subscription_id"])
	}
}

func TestDeleteBillingAttempt_404(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(404)
		_, _ = io.WriteString(w, `{"error":"not found"}`)
	})
	_, err := c.DeleteBillingAttempt(999, 999)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 error, got: %v", err)
	}
}

// ─── SkipBillingAttempt ───────────────────────────────────────────────────────

func TestSkipBillingAttempt_HappyPath(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription-billing-attempt/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			assertBodyField(t, body, "action", "skip")
			assertBodyFieldFloat(t, body, "id", 7)
			assertBodyFieldFloat(t, body, "subscription_id", 300)
		},
	))
	_, err := c.SkipBillingAttempt(7, 300)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSkipBillingAttempt_503(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(503)
	})
	_, err := c.SkipBillingAttempt(7, 300)
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Errorf("expected 503 error, got: %v", err)
	}
}

// ─── UnskipBillingAttempt ─────────────────────────────────────────────────────

func TestUnskipBillingAttempt_HappyPath(t *testing.T) {
	c := newTestClient(t, respondWithBodyCheck(
		t, http.MethodPut, "/subscription-billing-attempt/", 200, `{"success":true}`,
		func(t *testing.T, body map[string]any) {
			assertBodyField(t, body, "action", "unskip")
			assertBodyFieldFloat(t, body, "id", 8)
			assertBodyFieldFloat(t, body, "subscription_id", 400)
		},
	))
	_, err := c.UnskipBillingAttempt(8, 400)
	if err != nil {
		t.Fatal(err)
	}
}
