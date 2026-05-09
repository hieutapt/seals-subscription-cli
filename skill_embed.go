package main

import (
	_ "embed"

	"github.com/hieutapt/seals-subscription-cli/internal/skills"
)

//go:embed skills/seal-cli/SKILL.md
var embeddedSkillMD []byte

func init() {
	skills.SkillMD = embeddedSkillMD
}
