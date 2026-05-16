package testutil

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func TestAgentInspectionCommandsAvoidBooleanJSONFlags(t *testing.T) {
	t.Parallel()

	repoRoot := agentInspectionOutputGuardRepoRoot(t)
	targets := []string{
		filepath.Join(repoRoot, "internal", "cli", "agent", "status.go"),
		filepath.Join(repoRoot, "internal", "cli", "agent", "list.go"),
		filepath.Join(repoRoot, "internal", "cli", "agent", "catalog.go"),
		filepath.Join(repoRoot, "internal", "cli", "agent", "hooks.go"),
	}

	jsonFlagDecl := regexp.MustCompile(`BoolVarP?\([^)\n]*"json"`)

	for _, target := range targets {
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("read %s: %v", target, err)
		}
		if jsonFlagDecl.Find(data) != nil {
			t.Fatalf("%s reintroduces a boolean --json flag declaration in migrated agent inspection commands", target)
		}
	}
}

func agentInspectionOutputGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
