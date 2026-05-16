package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestP2PSkillsSpecMatchesPlaceholderOnlyEmbeddedState(t *testing.T) {
	t.Parallel()

	repoRoot := p2pSkillsSpecGuardRepoRoot(t)
	placeholder := filepath.Join(repoRoot, "skills", ".placeholder", "SKILL.md")
	if _, err := os.Stat(placeholder); err != nil {
		t.Fatalf("placeholder embedded skill scaffold missing: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(repoRoot, "skills", "p2p-*", "SKILL.md"))
	if err != nil {
		t.Fatalf("glob p2p embedded skills: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("unexpected embedded p2p skill files present: %v", matches)
	}

	specPath := filepath.Join(repoRoot, "openspec", "specs", "p2p-skills", "spec.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read %s: %v", specPath, err)
	}
	text := string(data)

	forbidden := []string{
		"The system SHALL provide 8 embedded skills for P2P operations",
		"skills/p2p-owner-shield/SKILL.md",
		"skills/p2p-reputation/SKILL.md",
		"skills/p2p-pricing/SKILL.md",
	}
	for _, needle := range forbidden {
		if strings.Contains(text, needle) {
			t.Fatalf("%s contains stale embedded-p2p-skill claim %q", specPath, needle)
		}
	}
}

func p2pSkillsSpecGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
