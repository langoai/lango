package x402

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestX402PackageHasNoContextTODORemaining(t *testing.T) {
	t.Parallel()

	pkgDir := x402PackageDir(t)
	needle := "context." + "TODO("

	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		t.Fatalf("read x402 package dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(pkgDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(data), needle) {
			t.Fatalf("%s contains forbidden %q", path, needle)
		}
	}
}

func TestCodebaseHasNoLegacyClientFactoryReferences(t *testing.T) {
	t.Parallel()

	repoRoot := x402RepoRoot(t)
	needle := "NewX402" + "Client"

	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".claude", ".codex":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(data), needle) {
			t.Fatalf("%s contains forbidden reference %q", path, needle)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan repo for %q: %v", needle, err)
	}
}

func TestX402PackageHasNoLegacyClientFactoryDefinition(t *testing.T) {
	t.Parallel()

	pkgDir := x402PackageDir(t)
	needle := "func " + "NewX402" + "Client("

	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		t.Fatalf("read x402 package dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(pkgDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(data), needle) {
			t.Fatalf("%s contains forbidden legacy definition %q", path, needle)
		}
	}
}

func x402PackageDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	return filepath.Dir(file)
}

func x402RepoRoot(t *testing.T) string {
	t.Helper()

	return filepath.Clean(filepath.Join(x402PackageDir(t), "..", ".."))
}
