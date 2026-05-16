package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestREADMEP2PGitSummariesStayTruthful(t *testing.T) {
	t.Parallel()

	repoRoot := p2pReadmeDocsGuardRepoRoot(t)
	target := filepath.Join(repoRoot, "README.md")

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read %s: %v", target, err)
	}

	text := string(data)
	forbidden := []string{
		"lango p2p git init <workspace-id>    Initialize a workspace git repository via the server-backed runtime",
		"lango p2p git push <workspace-id>    Push a workspace git bundle to peers via the server-backed runtime",
		"lango p2p git fetch <workspace-id>   Fetch a workspace git bundle from peers via the server-backed runtime",
	}
	for _, needle := range forbidden {
		if strings.Contains(text, needle) {
			t.Fatalf("%s contains stale README P2P git summary %q", target, needle)
		}
	}

	required := []string{
		"lango p2p git init <workspace-id>    Describe how to initialize a workspace git repository",
		"lango p2p git log <workspace-id>     Describe how to inspect workspace commit history",
		"lango p2p git diff <workspace-id> <from> <to>  Describe how to diff workspace commits",
		"lango p2p git push <workspace-id>    Describe how to push a workspace git bundle to peers",
		"lango p2p git fetch <workspace-id>   Describe how to fetch a workspace git bundle from peers",
	}
	for _, needle := range required {
		if !strings.Contains(text, needle) {
			t.Fatalf("%s must contain truthful README P2P git summary %q", target, needle)
		}
	}
}

func p2pReadmeDocsGuardRepoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
