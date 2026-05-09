package cmd

import (
	"github.com/hieutapt/seals-subscription-cli/internal/skills"
	"github.com/spf13/cobra"
)

func setupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Install the seal-cli agent skill into ~/.agents and ~/.claude",
		Long: `Install the bundled SKILL.md into your agent skill directories:

  ~/.agents/skills/seal-cli/SKILL.md   (universal agent store)
  ~/.claude/skills/seal-cli             (symlink, if Claude Code is installed)

This is run automatically by the install script. Re-running is safe — it
always overwrites with the latest version bundled in this binary.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return skills.InstallSkill()
		},
	}
}
