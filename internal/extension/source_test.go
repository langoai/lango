package extension

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFakePack builds a self-contained pack directory for tests.
func writeFakePack(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	manifest := `schema: lango.extension/v1
name: fake-pack
version: 0.1.0
description: Fake pack for tests
contents:
  skills:
    - name: foo
      path: skills/foo/SKILL.md
  prompts:
    - path: prompts/hello.md
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, manifestFileName), []byte(manifest), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "skills", "foo"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skills", "foo", "SKILL.md"),
		[]byte("---\nname: foo\nstatus: active\n---\nhello"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "prompts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "prompts", "hello.md"),
		[]byte("prompt body"), 0o644))
	return dir
}

func TestLocalSource_FetchHappyPath(t *testing.T) {
	t.Parallel()

	dir := writeFakePack(t)
	src := NewLocalSource(dir)
	wc, err := src.Fetch(context.Background())
	require.NoError(t, err)
	require.NotNil(t, wc)
	assert.Equal(t, "fake-pack", wc.Manifest.Name)
	assert.NotEmpty(t, wc.ManifestSHA256)
	assert.Len(t, wc.FileHashes, 2, "one per skill + one per prompt")
	// Cleanup is a no-op for local sources but must be non-nil.
	require.NoError(t, wc.Cleanup())
}

func TestLocalSource_MissingManifest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := NewLocalSource(dir)
	_, err := src.Fetch(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read manifest")
}

func TestLocalSource_NotADirectory(t *testing.T) {
	t.Parallel()

	f := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, os.WriteFile(f, []byte("x"), 0o644))
	src := NewLocalSource(f)
	_, err := src.Fetch(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestSplitGitRef(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in        string
		wantURL   string
		wantRef   string
	}{
		{"https://example.com/x.git", "https://example.com/x.git", ""},
		{"https://example.com/x.git#abc123", "https://example.com/x.git", "abc123"},
		{"git@example.com:user/x.git#main", "git@example.com:user/x.git", "main"},
	}
	for _, tt := range cases {
		t.Run(tt.in, func(t *testing.T) {
			url, ref := splitGitRef(tt.in)
			assert.Equal(t, tt.wantURL, url)
			assert.Equal(t, tt.wantRef, ref)
		})
	}
}

func TestFetchFromDir_HashesStableAcrossCalls(t *testing.T) {
	t.Parallel()

	dir := writeFakePack(t)
	w1, err := fetchFromDir(dir, dir, func() error { return nil })
	require.NoError(t, err)
	w2, err := fetchFromDir(dir, dir, func() error { return nil })
	require.NoError(t, err)
	assert.Equal(t, w1.ManifestSHA256, w2.ManifestSHA256)
	assert.Equal(t, w1.FileHashes, w2.FileHashes)
}

func TestLooksLikeSHA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give string
		want bool
	}{
		{"abc1234", true},         // 7 hex chars (minimum)
		{"abc1234def5678901234567890abcdef12345678", true}, // 40 hex chars (full SHA)
		{"abc123", false},         // too short (6)
		{"abc1234def5678901234567890abcdef123456789", false}, // too long (41)
		{"main", false},           // branch name
		{"v1.0.0", false},         // tag with dots
		{"ABCDEF1", false},        // uppercase hex
		{"abc123g", false},        // non-hex char
		{"", false},               // empty
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, looksLikeSHA(tt.give))
		})
	}
}

func TestGitSourceFetchSHA(t *testing.T) {
	t.Parallel()

	// Create a local bare repo with a commit, then clone using SHA reference.
	bareDir := t.TempDir()
	workDir := t.TempDir()

	// Init bare repo.
	runGit(t, bareDir, "init", "--bare")

	// Create a working clone, add a commit, push.
	runGit(t, "", "clone", bareDir, workDir)
	manifestContent := `schema: lango.extension/v1
name: sha-test
version: 0.1.0
description: SHA pin test
contents:
  skills:
    - name: sha-skill
      path: skills/sha-skill/SKILL.md
`
	require.NoError(t, os.WriteFile(filepath.Join(workDir, manifestFileName), []byte(manifestContent), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(workDir, "skills", "sha-skill"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workDir, "skills", "sha-skill", "SKILL.md"),
		[]byte("---\nname: sha-skill\nstatus: active\n---\nhi"), 0o644))
	runGit(t, workDir, "add", "-A")
	runGit(t, workDir, "commit", "-m", "init")
	runGit(t, workDir, "push", "origin", "HEAD")

	// Capture the commit SHA.
	sha := runGitOutput(t, workDir, "rev-parse", "HEAD")

	// Fetch using SHA pinning.
	src := NewGitSource(bareDir + "#" + sha)
	wc, err := src.Fetch(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = wc.Cleanup() })

	assert.Equal(t, "sha-test", wc.Manifest.Name)
	assert.Contains(t, wc.SourceRef, sha[:7])
}

// runGit executes a git command in the given dir. Empty dir uses cwd.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.com",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, string(out))
}

// runGitOutput executes a git command and returns trimmed stdout.
func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	require.NoError(t, err, "git %v", args)
	return strings.TrimSpace(string(out))
}

func TestFetchFromDir_PathEscapeRejected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outside := t.TempDir()
	target := filepath.Join(outside, "secret.md")
	require.NoError(t, os.WriteFile(target, []byte("x"), 0o644))

	manifest := `schema: lango.extension/v1
name: evil
version: 0.1.0
description: Should fail
contents:
  prompts:
    - path: escape.md
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, manifestFileName), []byte(manifest), 0o644))
	require.NoError(t, os.Symlink(target, filepath.Join(dir, "escape.md")))

	_, err := fetchFromDir(dir, dir, func() error { return nil })
	require.Error(t, err)
	assert.Contains(t, err.Error(), "escapes pack root")
}
