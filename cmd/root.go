package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hieutapt/seals-subscription-cli/internal/api"
	"github.com/hieutapt/seals-subscription-cli/internal/config"
)

var (
	profile     string
	jsonOut     bool
	versionInfo = "dev"
)

// SetVersionInfo is called from main with the ldflags values.
func SetVersionInfo(version, commit, date string) {
	versionInfo = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
}

// NewRootCmd builds and returns a fresh root command tree. It is called by
// Execute (production) and by tests (each test gets its own isolated tree).
func NewRootCmd() *cobra.Command {
	var localProfile string
	var localJSONOut bool

	root := &cobra.Command{
		Use:   "seal-cli",
		Short: "CLI for the Seal Subscriptions Merchant API",
		Long: `seal-cli wraps the Seal Subscriptions Merchant API.

Authentication:
  Set the SEAL_TOKEN environment variable, or configure a profile:
    seal-cli profile set --name default --token YOUR_TOKEN
`,
	}

	root.PersistentFlags().StringVar(&localProfile, "profile", "", "Config profile to use (default: current_profile in ~/.seal-cli.yaml)")
	root.PersistentFlags().BoolVar(&localJSONOut, "json", false, "Output raw JSON instead of pretty table")
	root.Version = versionInfo

	// Wire subcommands with their own closure over localProfile / localJSONOut.
	root.AddCommand(subscriptionCmd(&localProfile, &localJSONOut))
	root.AddCommand(billingAttemptCmd(&localProfile, &localJSONOut))
	root.AddCommand(profileCmd())

	return root
}

// Execute runs the CLI. Called from main.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

// newClient creates an API client for the active profile.
func newClient(profile string) (*api.Client, error) {
	token, err := config.ActiveToken(profile)
	if err != nil {
		return nil, err
	}
	return api.New(token), nil
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
