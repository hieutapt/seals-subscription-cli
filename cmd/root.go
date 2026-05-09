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

var rootCmd = &cobra.Command{
	Use:   "seal-cli",
	Short: "CLI for the Seal Subscriptions Merchant API",
	Long: `seal-cli wraps the Seal Subscriptions Merchant API.

Authentication:
  Set the SEAL_TOKEN environment variable, or configure a profile:
    seal-cli profile set --name default --token YOUR_TOKEN
`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&profile, "profile", "", "Config profile to use (default: current_profile in ~/.seal-cli.yaml)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "Output raw JSON instead of pretty table")
	rootCmd.Version = versionInfo

	rootCmd.AddCommand(subscriptionCmd())
	rootCmd.AddCommand(billingAttemptCmd())
	rootCmd.AddCommand(profileCmd())
}

// newClient creates an API client for the active profile.
func newClient() (*api.Client, error) {
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
