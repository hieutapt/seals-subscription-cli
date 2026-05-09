package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hieutapt/seals-subscription-cli/internal/config"
	"github.com/hieutapt/seals-subscription-cli/internal/output"
)

func profileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage named API profiles",
	}

	cmd.AddCommand(profileSetCmd())
	cmd.AddCommand(profileUseCmd())
	cmd.AddCommand(profileListCmd())

	return cmd
}

func profileSetCmd() *cobra.Command {
	var (
		name  string
		token string
		shop  string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Create or update a profile",
		Example: `  seal-cli profile set --name production --token abc123
  seal-cli profile set --name staging --token xyz --shop "My Staging Shop"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if token == "" {
				return fmt.Errorf("--token is required")
			}
			if err := config.SetProfile(name, token, shop); err != nil {
				return err
			}
			fmt.Printf("✓ Profile %q saved to ~/.seal-cli.yaml\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Profile name")
	cmd.Flags().StringVar(&token, "token", "", "Seal API token")
	cmd.Flags().StringVar(&shop, "shop", "", "Optional shop label")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("token")

	return cmd
}

func profileUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use [name]",
		Short: "Set the active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.UseProfile(args[0]); err != nil {
				return err
			}
			fmt.Printf("✓ Now using profile %q\n", args[0])
			return nil
		},
	}
}

func profileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.ListProfiles()
			if err != nil {
				return err
			}

			if len(cfg.Profiles) == 0 {
				fmt.Println("No profiles configured. Run: seal-cli profile set --name default --token YOUR_TOKEN")
				return nil
			}

			rows := [][]string{}
			for name, p := range cfg.Profiles {
				active := ""
				if name == cfg.CurrentProfile {
					active = "✓"
				}
				rows = append(rows, []string{name, p.Shop, active})
			}
			output.ProfileList([]string{"Profile", "Shop", "Active"}, rows)
			return nil
		},
	}
}
