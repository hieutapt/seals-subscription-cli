package skills

import (
	"fmt"
	"os"
	"path/filepath"
)

// SkillMD holds the embedded SKILL.md bytes. Populated at startup by
// skill_embed.go in package main (which can reach skills/seal-cli/SKILL.md).
var SkillMD []byte

const skillName = "seal-cli"

// InstallSkill writes the embedded SKILL.md to:
//   - ~/.agents/skills/seal-cli/SKILL.md  (universal agent store)
//   - ~/.claude/skills/seal-cli            (symlink → ../../.agents/skills/seal-cli)
//
// It is idempotent: existing files are silently overwritten with the latest
// embedded version. Agent root directories that do not exist are skipped.
func InstallSkill() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home dir: %w", err)
	}

	// 1. Write real skill file to the universal ~/.agents/skills/ store.
	agentsSkillDir := filepath.Join(home, ".agents", "skills", skillName)
	if err := os.MkdirAll(agentsSkillDir, 0755); err != nil {
		return fmt.Errorf("create ~/.agents/skills/%s: %w", skillName, err)
	}
	dest := filepath.Join(agentsSkillDir, "SKILL.md")
	if err := os.WriteFile(dest, SkillMD, 0644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}
	fmt.Printf("✓ ~/.agents/skills/%s/SKILL.md\n", skillName)

	// 2. Symlink ~/.claude/skills/seal-cli → ../../.agents/skills/seal-cli
	//    Only if ~/.claude/skills/ exists (Claude Code is installed).
	claudeSkillsDir := filepath.Join(home, ".claude", "skills")
	if _, statErr := os.Stat(claudeSkillsDir); statErr == nil {
		link := filepath.Join(claudeSkillsDir, skillName)
		// Relative target from the link's own directory.
		target := filepath.Join("..", "..", ".agents", "skills", skillName)

		// Remove any existing entry (file, dir, or stale symlink) before re-linking.
		_ = os.Remove(link)

		if err := os.Symlink(target, link); err != nil {
			fmt.Fprintf(os.Stderr, "  warning: could not symlink ~/.claude/skills/%s: %v\n", skillName, err)
		} else {
			fmt.Printf("✓ ~/.claude/skills/%s → ../../.agents/skills/%s\n", skillName, skillName)
		}
	}

	return nil
}
