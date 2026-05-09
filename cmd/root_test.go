package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// runCmd builds a fresh root command, sets the given args, captures stdout,
// and returns the output and any execution error. Each call gets an isolated
// command tree with no shared global state.
// Because our output package writes directly to os.Stdout, we redirect it at
// the OS level using a pipe.
func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Redirect os.Stdout so output.* writes are captured
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	root := NewRootCmd()
	root.SetOut(w) // also capture Cobra's own output (help, errors)
	root.SetErr(w)
	root.SetArgs(args)
	execErr := root.Execute()

	// Close the write end and drain the pipe
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	return buf.String(), execErr
}

// fakeServer starts an httptest.Server that responds to all requests with the
// given status code and JSON body. It sets SEAL_TOKEN and SEAL_API_BASE_URL so
// the cmd layer uses it instead of the real API. Cleaned up automatically.
func fakeServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("SEAL_TOKEN", "test-token")
	t.Setenv("SEAL_API_BASE_URL", srv.URL)
	return srv
}

// ─── Help / version smoke tests ───────────────────────────────────────────────

func TestRootCmd_Help(t *testing.T) {
	out, err := runCmd(t, "--help")
	if err != nil {
		t.Fatalf("--help returned error: %v", err)
	}
	if !strings.Contains(out, "seal-cli") {
		t.Errorf("help output missing 'seal-cli': %s", out)
	}
}

func TestRootCmd_UnknownCommand(t *testing.T) {
	_, err := runCmd(t, "nonexistent-command")
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
}

// ─── subscription get --agent ────────────────────────────────────────────────

// Realistic single-subscription payload the fake server returns.
const subDetailAPIResp = `{"success":true,"payload":{"id":42,"status":"ACTIVE","email":"jane@example.com","first_name":"Jane","last_name":"Doe","currency":"USD","total_value":29.0,"delivery_interval":"1 month","billing_interval":"1 month","s_address1":"","s_city":"","s_country":"","b_address1":"","b_city":"","b_country":"","card_brand":"visa","card_last_digits":"4242","card_expiry_month":"12","card_expiry_year":"2027","items":[],"billing_attempts":[]}}`

func TestSubscriptionGet_AgentFlag(t *testing.T) {
	fakeServer(t, 200, subDetailAPIResp)

	out, err := runCmd(t, "subscription", "get", "42", "--agent")
	if err != nil {
		t.Fatalf("--agent returned error: %v", err)
	}
	// Must be a single compact JSON line
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("--agent output: expected 1 line, got %d:\n%s", len(lines), out)
	}
	var m map[string]any
	if jsonErr := json.Unmarshal([]byte(lines[0]), &m); jsonErr != nil {
		t.Fatalf("--agent output is not valid JSON: %v\noutput: %q", jsonErr, out)
	}
	if m["id"] != float64(42) {
		t.Errorf("id: got %v, want 42", m["id"])
	}
	if m["status"] != "ACTIVE" {
		t.Errorf("status: got %v, want ACTIVE", m["status"])
	}
}

func TestSubscriptionList_AgentFlag(t *testing.T) {
	const listResp = `{"success":true,"payload":{"page":1,"total_pages":1,"subscriptions":[{"id":1,"status":"ACTIVE","email":"a@x.com","first_name":"A","last_name":"X","currency":"USD","total_value":10,"delivery_interval":"1 month"}]}}`
	fakeServer(t, 200, listResp)

	out, err := runCmd(t, "subscription", "list", "--agent")
	if err != nil {
		t.Fatalf("--agent list returned error: %v", err)
	}
	// Must be NDJSON: 1 sub line + 1 _meta line
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("--agent list: expected 2 lines, got %d:\n%s", len(lines), out)
	}
	// Last line must have _meta
	var last map[string]any
	_ = json.Unmarshal([]byte(lines[len(lines)-1]), &last)
	if _, ok := last["_meta"]; !ok {
		t.Errorf("last line missing _meta: %s", lines[len(lines)-1])
	}
}

func TestSubscriptionCancel_AgentFlag(t *testing.T) {
	fakeServer(t, 200, `{"success":true,"message":"Subscription cancelled."}`)

	out, err := runCmd(t, "subscription", "cancel", "42", "--agent")
	if err != nil {
		t.Fatalf("--agent cancel returned error: %v", err)
	}
	var m map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); jsonErr != nil {
		t.Fatalf("--agent cancel output not valid JSON: %v\noutput: %q", jsonErr, out)
	}
	if m["ok"] != true {
		t.Errorf("ok: got %v, want true", m["ok"])
	}
}

func TestAgent_WinsOverJSON(t *testing.T) {
	fakeServer(t, 200, subDetailAPIResp)

	out, err := runCmd(t, "subscription", "get", "42", "--agent", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// --agent wins: output must be a single compact line, not pretty-printed
	if strings.Contains(out, "\n  ") {
		t.Errorf("--agent should win over --json but got pretty-printed output:\n%s", out)
	}
}

func TestSEAL_AGENT_MODE_EnvVar(t *testing.T) {
	fakeServer(t, 200, subDetailAPIResp)
	t.Setenv("SEAL_AGENT_MODE", "1")

	out, err := runCmd(t, "subscription", "get", "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should behave like --agent: single compact JSON line
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("SEAL_AGENT_MODE=1: expected 1 line, got %d:\n%s", len(lines), out)
	}
	var m map[string]any
	if jsonErr := json.Unmarshal([]byte(lines[0]), &m); jsonErr != nil {
		t.Fatalf("SEAL_AGENT_MODE=1 output not valid JSON: %v\noutput: %q", jsonErr, out)
	}
}

// ─── subscription get --json ──────────────────────────────────────────────────

func TestSubscriptionGet_JSONFlag(t *testing.T) {
	const apiResp = `{"id":42,"status":"active"}`
	fakeServer(t, 200, apiResp)

	out, err := runCmd(t, "subscription", "get", "42", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Output must be valid JSON containing the id field
	var m map[string]any
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); jsonErr != nil {
		t.Fatalf("--json output is not valid JSON: %v\noutput: %q", jsonErr, out)
	}
	if m["id"] != float64(42) {
		t.Errorf("id: got %v, want 42", m["id"])
	}
}

// ─── SEAL_TOKEN missing → clear error ────────────────────────────────────────

func TestSubscriptionGet_MissingToken(t *testing.T) {
	// No SEAL_TOKEN, no config file
	t.Setenv("SEAL_TOKEN", "")
	t.Setenv("HOME", t.TempDir())

	_, err := runCmd(t, "subscription", "get", "1")
	if err == nil {
		t.Fatal("expected error when token is missing, got nil")
	}
}

// ─── subscription list alias ──────────────────────────────────────────────────

func TestSubscriptionList_ViaAlias(t *testing.T) {
	fakeServer(t, 200, `{"subscriptions":[]}`)

	// "s ls" should work identically to "subscription list"
	_, err := runCmd(t, "s", "ls", "--json")
	if err != nil {
		t.Fatalf("alias 's ls' returned error: %v", err)
	}
}

// ─── billing-attempt alias ────────────────────────────────────────────────────

func TestBillingAttempt_ViaAlias(t *testing.T) {
	fakeServer(t, 200, `{"success":true}`)

	_, err := runCmd(t, "ba", "skip", "--id", "1", "--subscription-id", "100")
	if err != nil {
		t.Fatalf("alias 'ba skip' returned error: %v", err)
	}
}

// ─── profile subcommands (local only, no HTTP) ────────────────────────────────

func TestProfileSet_And_List(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SEAL_TOKEN", "")

	_, err := runCmd(t, "profile", "set", "--name", "dev", "--token", "devtoken")
	if err != nil {
		t.Fatalf("profile set: %v", err)
	}

	out, err := runCmd(t, "profile", "list")
	if err != nil {
		t.Fatalf("profile list: %v", err)
	}
	if !strings.Contains(out, "dev") {
		t.Errorf("profile list output missing 'dev': %s", out)
	}
}
