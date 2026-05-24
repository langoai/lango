package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAgentMemorySpecAvoidsDeletedAppToolBuilderPathClaim(t *testing.T) {
	t.Parallel()

	repoRoot := agentMemorySpecGuardRepoRoot(t)
	specPath := filepath.Join(repoRoot, "openspec", "specs", "agent-memory", "spec.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read %s: %v", specPath, err)
	}
	text := string(data)

	if strings.Contains(text, "internal/app/tools_agentmemory.go") {
		t.Fatalf("%s contains stale deleted app-local builder path", specPath)
	}
	if !strings.Contains(text, "internal/agentmemory/tools.go") && !strings.Contains(text, "current app module wiring") {
		t.Fatalf("%s does not describe the current agent memory builder ownership clearly", specPath)
	}
}

func agentMemorySpecGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
