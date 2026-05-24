package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadResidualErrorBranches(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tool := New(Config{AllowedPaths: []string{tmpDir}})

	_, err := tool.Read(filepath.Join(tmpDir, "missing.txt"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "file not found")

	_, err = tool.Read(tmpDir)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot read directory")
}

func TestReadLinesResidualBranches(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tool := New(Config{AllowedPaths: []string{tmpDir}})

	_, err := tool.ReadLines(filepath.Join(tmpDir, "missing.txt"), 1, 1)
	require.Error(t, err)
	assert.ErrorContains(t, err, "open file")

	shortFile := filepath.Join(tmpDir, "short.txt")
	require.NoError(t, os.WriteFile(shortFile, []byte("one\ntwo"), 0644))
	got, err := tool.ReadLines(shortFile, 3, 10)
	require.NoError(t, err)
	assert.Empty(t, got)

	longLineFile := filepath.Join(tmpDir, "long-line.txt")
	require.NoError(t, os.WriteFile(longLineFile, []byte(strings.Repeat("x", 70*1024)), 0644))
	_, err = tool.ReadLines(longLineFile, 1, 1)
	require.Error(t, err)
	assert.ErrorContains(t, err, "read error")
}

func TestEditResidualErrorBranches(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tool := New(Config{AllowedPaths: []string{tmpDir}})

	editable := filepath.Join(tmpDir, "editable.txt")
	require.NoError(t, os.WriteFile(editable, []byte("one\ntwo"), 0644))

	tests := []struct {
		name        string
		path        string
		startLine   int
		endLine     int
		wantErrText string
	}{
		{
			name:        "open missing file",
			path:        filepath.Join(tmpDir, "missing.txt"),
			startLine:   1,
			endLine:     1,
			wantErrText: "open file",
		},
		{
			name:        "invalid start before first line",
			path:        editable,
			startLine:   0,
			endLine:     1,
			wantErrText: "invalid start line",
		},
		{
			name:        "invalid start beyond append position",
			path:        editable,
			startLine:   4,
			endLine:     4,
			wantErrText: "invalid start line",
		},
		{
			name:        "end before start",
			path:        editable,
			startLine:   2,
			endLine:     1,
			wantErrText: "end line must be >= start line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.Edit(tt.path, tt.startLine, tt.endLine, "replacement")
			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErrText)
		})
	}

	longLineFile := filepath.Join(tmpDir, "long-edit.txt")
	require.NoError(t, os.WriteFile(longLineFile, []byte(strings.Repeat("x", 70*1024)), 0644))
	err := tool.Edit(longLineFile, 1, 1, "replacement")
	require.Error(t, err)
	assert.ErrorContains(t, err, "read error")
}

func TestListDirResidualErrorBranches(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tool := New(Config{AllowedPaths: []string{tmpDir}})

	filePath := filepath.Join(tmpDir, "not-a-dir.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("file"), 0644))

	_, err := tool.ListDir(filePath)
	require.Error(t, err)
	assert.ErrorContains(t, err, "read directory")

	_, err = tool.ListDir(filepath.Join(tmpDir, "missing"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "read directory")
}

func TestWriteResidualValidationParentAndRenameBranches(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	allowedDir := filepath.Join(tmpDir, "allowed")
	blockedDir := filepath.Join(allowedDir, "blocked")
	require.NoError(t, os.MkdirAll(allowedDir, 0755))

	tool := New(Config{
		AllowedPaths: []string{allowedDir},
		BlockedPaths: []string{blockedDir},
	})

	err := tool.Write(filepath.Join(blockedDir, "secret.txt"), "secret")
	require.Error(t, err)
	assert.ErrorContains(t, err, "access denied: protected path")

	nestedFile := filepath.Join(allowedDir, "nested", "leaf", "note.txt")
	require.NoError(t, tool.Write(nestedFile, "created"))
	got, err := os.ReadFile(nestedFile)
	require.NoError(t, err)
	assert.Equal(t, "created", string(got))

	existingDir := filepath.Join(allowedDir, "existing-dir")
	require.NoError(t, os.MkdirAll(existingDir, 0755))
	err = tool.Write(existingDir, "content")
	require.Error(t, err)
	assert.ErrorContains(t, err, "rename file")
	assert.NoFileExists(t, existingDir+".tmp")
}

func TestReadWithMetaResidualErrorBranches(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	tool := New(Config{AllowedPaths: []string{tmpDir}, MaxReadSize: 8})

	_, err := tool.ReadWithMeta(filepath.Join(tmpDir, "missing.txt"), 0, 0)
	require.Error(t, err)
	assert.ErrorContains(t, err, "file not found")

	_, err = tool.ReadWithMeta(tmpDir, 0, 0)
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot read directory")

	largeFile := filepath.Join(tmpDir, "large.txt")
	require.NoError(t, os.WriteFile(largeFile, []byte("larger than eight bytes"), 0644))
	_, err = tool.ReadWithMeta(largeFile, 0, 0)
	require.Error(t, err)
	assert.ErrorContains(t, err, "file too large")

	longLineFile := filepath.Join(tmpDir, "long-line.txt")
	require.NoError(t, os.WriteFile(longLineFile, []byte(strings.Repeat("x", 70*1024)), 0644))
	longLineTool := New(Config{AllowedPaths: []string{tmpDir}, MaxReadSize: 100 * 1024})
	_, err = longLineTool.ReadWithMeta(longLineFile, 0, 0)
	require.Error(t, err)
	assert.ErrorContains(t, err, "read file")
}

func TestCountLinesOpenError(t *testing.T) {
	t.Parallel()

	_, err := countLines(filepath.Join(t.TempDir(), "missing.txt"))
	require.Error(t, err)
}

func TestValidatePathTraversalBrokenSymlinkAndAccessEdges(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	allowedDir := filepath.Join(tmpDir, "allowed")
	blockedDir := filepath.Join(tmpDir, "blocked")
	require.NoError(t, os.MkdirAll(allowedDir, 0755))
	require.NoError(t, os.MkdirAll(blockedDir, 0755))

	traversalTool := New(Config{})
	_, err := traversalTool.validatePath("safe/../still/../../escape.txt")
	require.Error(t, err)
	assert.ErrorContains(t, err, "path traversal not allowed")

	brokenLink := filepath.Join(allowedDir, "broken-link")
	requireSymlink(t, filepath.Join(tmpDir, "missing-target"), brokenLink)
	tool := New(Config{
		AllowedPaths: []string{allowedDir},
		BlockedPaths: []string{blockedDir},
	})
	got, err := tool.validatePath(brokenLink)
	require.NoError(t, err)
	assert.Equal(t, brokenLink, got)

	allowedNested := filepath.Join(allowedDir, "nested", "edge.txt")
	require.NoError(t, os.MkdirAll(filepath.Dir(allowedNested), 0755))
	require.NoError(t, os.WriteFile(allowedNested, []byte("allowed"), 0644))
	wantAllowedNested, err := filepath.EvalSymlinks(allowedNested)
	require.NoError(t, err)
	got, err = tool.validatePath(allowedNested)
	require.NoError(t, err)
	assert.Equal(t, wantAllowedNested, got)

	_, err = tool.validatePath(filepath.Join(blockedDir, "edge.txt"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "access denied: protected path")
}
