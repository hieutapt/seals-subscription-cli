package config

import (
	"os"
	"testing"
)

// isolate points HOME at a fresh temp dir and clears SEAL_TOKEN so every test
// starts from a clean slate. It restores both automatically via t.Cleanup.
func isolate(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SEAL_TOKEN", "")
}

// ─── Load ─────────────────────────────────────────────────────────────────────

func TestLoad_MissingFile_ReturnsEmptyConfig(t *testing.T) {
	isolate(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.CurrentProfile != "default" {
		t.Errorf("CurrentProfile: got %q, want %q", cfg.CurrentProfile, "default")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("Profiles: expected empty map, got %v", cfg.Profiles)
	}
}

// ─── SetProfile + Load round-trip ─────────────────────────────────────────────

func TestSetProfile_PersistsAndLoads(t *testing.T) {
	isolate(t)
	if err := SetProfile("prod", "tok-abc", "My Shop"); err != nil {
		t.Fatalf("SetProfile: %v", err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	p, ok := cfg.Profiles["prod"]
	if !ok {
		t.Fatal("profile 'prod' not found after SetProfile")
	}
	if p.Token != "tok-abc" {
		t.Errorf("Token: got %q, want tok-abc", p.Token)
	}
	if p.Shop != "My Shop" {
		t.Errorf("Shop: got %q, want 'My Shop'", p.Shop)
	}
}

func TestSetProfile_OverwritesExisting(t *testing.T) {
	isolate(t)
	_ = SetProfile("staging", "old-token", "")
	_ = SetProfile("staging", "new-token", "Updated Shop")

	cfg, _ := Load()
	p := cfg.Profiles["staging"]
	if p.Token != "new-token" {
		t.Errorf("Token: got %q, want new-token", p.Token)
	}
	if p.Shop != "Updated Shop" {
		t.Errorf("Shop: got %q, want 'Updated Shop'", p.Shop)
	}
}

func TestSetProfile_MultipleProfilesCoexist(t *testing.T) {
	isolate(t)
	_ = SetProfile("alpha", "tok-a", "")
	_ = SetProfile("beta", "tok-b", "")

	cfg, _ := Load()
	if _, ok := cfg.Profiles["alpha"]; !ok {
		t.Error("alpha profile missing")
	}
	if _, ok := cfg.Profiles["beta"]; !ok {
		t.Error("beta profile missing")
	}
}

// ─── UseProfile ───────────────────────────────────────────────────────────────

func TestUseProfile_SetsCurrentProfile(t *testing.T) {
	isolate(t)
	_ = SetProfile("work", "tok-w", "")
	if err := UseProfile("work"); err != nil {
		t.Fatalf("UseProfile: %v", err)
	}
	cfg, _ := Load()
	if cfg.CurrentProfile != "work" {
		t.Errorf("CurrentProfile: got %q, want work", cfg.CurrentProfile)
	}
}

func TestUseProfile_NonExistentReturnsError(t *testing.T) {
	isolate(t)
	err := UseProfile("ghost")
	if err == nil {
		t.Fatal("expected error for non-existent profile, got nil")
	}
}

// ─── ActiveToken precedence ───────────────────────────────────────────────────

func TestActiveToken_EnvVarTakesPrecedence(t *testing.T) {
	isolate(t)
	_ = SetProfile("default", "profile-token", "")
	t.Setenv("SEAL_TOKEN", "env-token")

	got, err := ActiveToken("")
	if err != nil {
		t.Fatalf("ActiveToken: %v", err)
	}
	if got != "env-token" {
		t.Errorf("got %q, want env-token", got)
	}
}

func TestActiveToken_UsesCurrentProfile(t *testing.T) {
	isolate(t)
	_ = SetProfile("myprofile", "my-token", "")
	_ = UseProfile("myprofile")

	got, err := ActiveToken("")
	if err != nil {
		t.Fatalf("ActiveToken: %v", err)
	}
	if got != "my-token" {
		t.Errorf("got %q, want my-token", got)
	}
}

func TestActiveToken_ExplicitProfileNameOverridesCurrentProfile(t *testing.T) {
	isolate(t)
	_ = SetProfile("current", "current-token", "")
	_ = SetProfile("other", "other-token", "")
	_ = UseProfile("current")

	got, err := ActiveToken("other")
	if err != nil {
		t.Fatalf("ActiveToken: %v", err)
	}
	if got != "other-token" {
		t.Errorf("got %q, want other-token", got)
	}
}

func TestActiveToken_MissingProfileReturnsError(t *testing.T) {
	isolate(t)
	_, err := ActiveToken("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing profile, got nil")
	}
}

func TestActiveToken_EmptyTokenReturnsError(t *testing.T) {
	isolate(t)
	// Manually save a profile with an empty token
	cfg, _ := Load()
	cfg.Profiles["broken"] = Profile{Token: ""}
	_ = cfg.Save()
	cfg.CurrentProfile = "broken"
	_ = cfg.Save()

	_, err := ActiveToken("broken")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

// ─── Config file permissions ──────────────────────────────────────────────────

func TestSave_FileMode0600(t *testing.T) {
	isolate(t)
	_ = SetProfile("default", "tok", "")

	path, _ := configPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file permissions: got %04o, want 0600", perm)
	}
}

// ─── ListProfiles ─────────────────────────────────────────────────────────────

func TestListProfiles_ReturnsAllProfiles(t *testing.T) {
	isolate(t)
	_ = SetProfile("a", "tok-a", "Shop A")
	_ = SetProfile("b", "tok-b", "Shop B")

	cfg, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}
	if len(cfg.Profiles) != 2 {
		t.Errorf("expected 2 profiles, got %d", len(cfg.Profiles))
	}
}
