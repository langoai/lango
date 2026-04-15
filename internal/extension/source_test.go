package extension

import (
	"context"
	"os"
	"path/filepath"
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
