package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultProfile = "default"

// Config holds all profiles.
type Config struct {
	CurrentProfile string             `yaml:"current_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
}

// Profile holds per-shop settings.
type Profile struct {
	Token string `yaml:"token"`
	Shop  string `yaml:"shop,omitempty"` // optional label
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".seal-cli.yaml"), nil
}

// Load reads the config file. Returns empty config if not found.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	cfg := &Config{
		CurrentProfile: defaultProfile,
		Profiles:       make(map[string]Profile),
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	return cfg, nil
}

// Save writes the config to disk.
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// ActiveToken returns the token for the active profile, falling back to SEAL_TOKEN env var.
func ActiveToken(profileName string) (string, error) {
	// Env var takes precedence if explicitly set
	if t := os.Getenv("SEAL_TOKEN"); t != "" {
		return t, nil
	}

	cfg, err := Load()
	if err != nil {
		return "", err
	}

	name := profileName
	if name == "" {
		name = cfg.CurrentProfile
	}
	if name == "" {
		name = defaultProfile
	}

	p, ok := cfg.Profiles[name]
	if !ok {
		return "", fmt.Errorf("profile %q not found in config. Set SEAL_TOKEN env var or run: seal-cli profile set --name %s --token YOUR_TOKEN", name, name)
	}
	if p.Token == "" {
		return "", fmt.Errorf("profile %q has no token set", name)
	}
	return p.Token, nil
}

// SetProfile sets or updates a named profile.
func SetProfile(name, token, shop string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.Profiles[name] = Profile{Token: token, Shop: shop}
	return cfg.Save()
}

// UseProfile sets the current active profile.
func UseProfile(name string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	if _, ok := cfg.Profiles[name]; !ok {
		return fmt.Errorf("profile %q does not exist", name)
	}
	cfg.CurrentProfile = name
	return cfg.Save()
}

// ListProfiles returns the config for display.
func ListProfiles() (*Config, error) {
	return Load()
}
