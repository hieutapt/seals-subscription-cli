package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallSkill_WritesSkillMD(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	SkillMD = []byte("# seal-cli skill")

	if err := InstallSkill(); err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	dest := filepath.Join(home, ".agents", "skills", skillName, "SKILL.md")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("expected SKILL.md at %s: %v", dest, err)
	}
	if string(got) != string(SkillMD) {
		t.Errorf("SKILL.md content mismatch: got %q, want %q", got, SkillMD)
	}
}

func TestInstallSkill_Idempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	SkillMD = []byte("version 1")
	if err := InstallSkill(); err != nil {
		t.Fatalf("first InstallSkill() error: %v", err)
	}

	SkillMD = []byte("version 2")
	if err := InstallSkill(); err != nil {
		t.Fatalf("second InstallSkill() error: %v", err)
	}

	dest := filepath.Join(home, ".agents", "skills", skillName, "SKILL.md")
	got, _ := os.ReadFile(dest)
	if string(got) != "version 2" {
		t.Errorf("expected overwrite with version 2, got %q", got)
	}
}

func TestInstallSkill_NoClaudeDir_NoError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// ~/.claude/skills does NOT exist — should not error

	SkillMD = []byte("# skill")
	if err := InstallSkill(); err != nil {
		t.Fatalf("InstallSkill() should not error when ~/.claude/skills missing: %v", err)
	}

	// No symlink should be created
	link := filepath.Join(home, ".claude", "skills", skillName)
	if _, err := os.Lstat(link); err == nil {
		t.Errorf("expected no symlink at %s, but it exists", link)
	}
}

func TestInstallSkill_CreatesClaudeSymlink(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Create ~/.claude/skills so the symlink branch runs
	claudeSkillsDir := filepath.Join(home, ".claude", "skills")
	if err := os.MkdirAll(claudeSkillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	SkillMD = []byte("# skill")
	if err := InstallSkill(); err != nil {
		t.Fatalf("InstallSkill() error: %v", err)
	}

	link := filepath.Join(claudeSkillsDir, skillName)
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("expected symlink at %s: %v", link, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected a symlink at %s, got mode %v", link, info.Mode())
	}

	target, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("..", "..", ".agents", "skills", skillName)
	if target != want {
		t.Errorf("symlink target = %q, want %q", target, want)
	}
}

func TestInstallSkill_SymlinkIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	claudeSkillsDir := filepath.Join(home, ".claude", "skills")
	if err := os.MkdirAll(claudeSkillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	SkillMD = []byte("# skill")
	// Run twice — second call should replace the symlink without error
	for i := 0; i < 2; i++ {
		if err := InstallSkill(); err != nil {
			t.Fatalf("InstallSkill() call %d error: %v", i+1, err)
		}
	}

	link := filepath.Join(claudeSkillsDir, skillName)
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("symlink missing after second install: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected symlink, got %v", info.Mode())
	}
}
