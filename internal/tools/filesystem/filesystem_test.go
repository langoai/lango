package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadWrite(t *testing.T) {
	t.Parallel()

	tool := New(Config{})
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write
	content := "hello\nworld"
	require.NoError(t, tool.Write(testFile, content))

	// Read
	result, err := tool.Read(testFile)
	require.NoError(t, err)
	assert.Equal(t, content, result)
}

func TestReadLines(t *testing.T) {
	t.Parallel()

	tool := New(Config{})
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "lines.txt")

	content := "line1\nline2\nline3\nline4\nline5"
	require.NoError(t, tool.Write(testFile, content))

	result, err := tool.ReadLines(testFile, 2, 4)
	require.NoError(t, err)
	assert.Equal(t, "line2\nline3\nline4", result)
}

func TestEdit(t *testing.T) {
	t.Parallel()

	tool := New(Config{})
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "edit.txt")

	content := "line1\nold\nline3"
	require.NoError(t, tool.Write(testFile, content))
	require.NoError(t, tool.Edit(testFile, 2, 2, "new"))

	result, _ := tool.Read(testFile)
	assert.Equal(t, "line1\nnew\nline3", result)
}

func TestListDir(t *testing.T) {
	t.Parallel()

	tool := New(Config{})
	tmpDir := t.TempDir()

	// Create some files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("b"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	files, err := tool.ListDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 3)
}

func TestBuildTools_FsListKeepsPathOptionalAndDefaultsToDot(t *testing.T) {
	t.Parallel()

	tool := New(Config{})
	tools := BuildTools(tool)

	var listTool *agent.Tool
	for _, candidate := range tools {
		if candidate.Name == "fs_list" {
			listTool = candidate
			break
		}
	}
	require.NotNil(t, listTool)

	required, ok := listTool.Parameters["required"].([]string)
	if ok {
		assert.NotContains(t, required, "path")
	}

	cwd := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(cwd, "visible.txt"), []byte("x"), 0644))
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(cwd))
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	got, err := listTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)

	entries, ok := got.([]FileInfo)
	require.True(t, ok)
	require.NotEmpty(t, entries)
	assert.Contains(t, entries[0].Path, cwd)
}

func TestBuildTools_FsWriteAndFsEditRequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tool := New(Config{})
	tools := BuildTools(tool)

	var writeTool *agent.Tool
	var editTool *agent.Tool
	for _, candidate := range tools {
		switch candidate.Name {
		case "fs_write":
			writeTool = candidate
		case "fs_edit":
			editTool = candidate
		}
	}
	require.NotNil(t, writeTool)
	require.NotNil(t, editTool)

	writeResult, err := writeTool.Handler(context.Background(), map[string]interface{}{
		"path": "tmp.txt",
	})
	require.Error(t, err)
	assert.Nil(t, writeResult)
	assert.ErrorContains(t, err, "missing content parameter")

	editResult, err := editTool.Handler(context.Background(), map[string]interface{}{
		"path":      "tmp.txt",
		"startLine": float64(1),
		"content":   "patched",
	})
	require.Error(t, err)
	assert.Nil(t, editResult)
	assert.ErrorContains(t, err, "missing endLine parameter")
}

func TestPathValidation(t *testing.T) {
	t.Parallel()

	tool := New(Config{
		AllowedPaths: []string{"/tmp/allowed"},
	})

	// Should fail for paths outside allowed
	_, err := tool.validatePath("/etc/passwd")
	require.Error(t, err)
}

func TestFileSizeLimit(t *testing.T) {
	t.Parallel()

	tool := New(Config{MaxReadSize: 10})
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Write file larger than limit
	os.WriteFile(testFile, []byte("this is larger than 10 bytes"), 0644)

	_, err := tool.Read(testFile)
	require.Error(t, err)
}

func TestStat(t *testing.T) {
	t.Parallel()

	tool := New(Config{})
	tmpDir := t.TempDir()

	tests := []struct {
		give      string
		setup     func(t *testing.T) string
		wantErr   bool
		wantLines int
		wantIsDir bool
	}{
		{
			give: "regular file",
			setup: func(t *testing.T) string {
				p := filepath.Join(tmpDir, "stat_regular.txt")
				require.NoError(t, os.WriteFile(p, []byte("line1\nline2\nline3"), 0644))
				return p
			},
			wantLines: 3,
			wantIsDir: false,
		},
		{
			give: "directory",
			setup: func(t *testing.T) string {
				p := filepath.Join(tmpDir, "stat_dir")
				require.NoError(t, os.MkdirAll(p, 0755))
				return p
			},
			wantLines: 0,
			wantIsDir: true,
		},
		{
			give: "non-existent file",
			setup: func(t *testing.T) string {
				return filepath.Join(tmpDir, "does_not_exist.txt")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			path := tt.setup(t)
			result, err := tool.Stat(path)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantLines, result.Lines)
			assert.Equal(t, tt.wantIsDir, result.IsDir)
			assert.NotZero(t, result.ModTime)
			assert.NotEmpty(t, result.Permission)

			if !tt.wantIsDir {
				assert.Greater(t, result.Size, int64(0))
			}
		})
	}
}

func TestReadWithMeta(t *testing.T) {
	t.Parallel()

	tool := New(Config{})
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "readmeta.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("line1\nline2\nline3\nline4\nline5"), 0644))

	tests := []struct {
		give        string
		giveOffset  int
		giveLimit   int
		wantContent string
		wantTotal   int
		wantOffset  int
		wantLimit   int
	}{
		{
			give:        "full read offset=0 limit=0",
			giveOffset:  0,
			giveLimit:   0,
			wantContent: "line1\nline2\nline3\nline4\nline5",
			wantTotal:   5,
			wantOffset:  1,
			wantLimit:   0,
		},
		{
			give:        "with offset",
			giveOffset:  3,
			giveLimit:   0,
			wantContent: "line3\nline4\nline5",
			wantTotal:   5,
			wantOffset:  3,
			wantLimit:   0,
		},
		{
			give:        "with limit",
			giveOffset:  0,
			giveLimit:   2,
			wantContent: "line1\nline2",
			wantTotal:   5,
			wantOffset:  1,
			wantLimit:   2,
		},
		{
			give:        "offset and limit combined",
			giveOffset:  2,
			giveLimit:   2,
			wantContent: "line2\nline3",
			wantTotal:   5,
			wantOffset:  2,
			wantLimit:   2,
		},
		{
			give:        "large offset beyond file",
			giveOffset:  100,
			giveLimit:   0,
			wantContent: "",
			wantTotal:   5,
			wantOffset:  100,
			wantLimit:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			result, err := tool.ReadWithMeta(testFile, tt.giveOffset, tt.giveLimit)
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, result.Content)
			assert.Equal(t, tt.wantTotal, result.TotalLines)
			assert.Equal(t, tt.wantOffset, result.Offset)
			assert.Equal(t, tt.wantLimit, result.Limit)
			assert.Greater(t, result.Size, int64(0))
		})
	}
}

func TestExistsReturnsTrueFalseAndValidationErrors(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	allowedDir := filepath.Join(tmpDir, "allowed")
	blockedDir := filepath.Join(allowedDir, "blocked")
	require.NoError(t, os.MkdirAll(blockedDir, 0755))
	existingFile := filepath.Join(allowedDir, "exists.txt")
	require.NoError(t, os.WriteFile(existingFile, []byte("present"), 0644))
	blockedFile := filepath.Join(blockedDir, "secret.txt")
	require.NoError(t, os.WriteFile(blockedFile, []byte("secret"), 0644))

	tool := New(Config{
		AllowedPaths: []string{allowedDir},
		BlockedPaths: []string{blockedDir},
	})

	exists, err := tool.Exists(existingFile)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = tool.Exists(filepath.Join(allowedDir, "missing.txt"))
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = tool.Exists(blockedFile)
	require.Error(t, err)
	assert.False(t, exists)
	assert.ErrorContains(t, err, "access denied: protected path")
}

func TestMkdirCreatesAllowedPathAndRejectsBlockedPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	allowedDir := filepath.Join(tmpDir, "allowed")
	blockedDir := filepath.Join(allowedDir, "blocked")
	require.NoError(t, os.MkdirAll(allowedDir, 0755))

	tool := New(Config{
		AllowedPaths: []string{allowedDir},
		BlockedPaths: []string{blockedDir},
	})

	createdDir := filepath.Join(allowedDir, "nested", "leaf")
	require.NoError(t, tool.Mkdir(createdDir))
	info, err := os.Stat(createdDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	err = tool.Mkdir(filepath.Join(blockedDir, "secret"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "access denied: protected path")
	assert.NoDirExists(t, blockedDir)
}

func TestCopyCopiesAllowedFileAndRejectsMissingOrBlockedPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	allowedDir := filepath.Join(tmpDir, "allowed")
	blockedDir := filepath.Join(allowedDir, "blocked")
	require.NoError(t, os.MkdirAll(allowedDir, 0755))
	require.NoError(t, os.MkdirAll(blockedDir, 0755))
	srcFile := filepath.Join(allowedDir, "src.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("copy me"), 0644))

	tool := New(Config{
		AllowedPaths: []string{allowedDir},
		BlockedPaths: []string{blockedDir},
	})

	dstFile := filepath.Join(allowedDir, "nested", "dst.txt")
	require.NoError(t, tool.Copy(srcFile, dstFile))
	got, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, "copy me", string(got))

	err = tool.Copy(filepath.Join(allowedDir, "missing.txt"), filepath.Join(allowedDir, "missing-dst.txt"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "open source")

	err = tool.Copy(srcFile, filepath.Join(blockedDir, "dst.txt"))
	require.Error(t, err)
	assert.ErrorContains(t, err, "access denied: protected path")
}

func TestDeleteHandlesLocalRecursiveP2PRestrictedAndSymlinkRemoval(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	allowedDir := filepath.Join(tmpDir, "allowed")
	require.NoError(t, os.MkdirAll(allowedDir, 0755))
	tool := New(Config{AllowedPaths: []string{allowedDir}})

	localDir := filepath.Join(allowedDir, "local-dir")
	require.NoError(t, os.MkdirAll(localDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(localDir, "child.txt"), []byte("child"), 0644))
	require.NoError(t, tool.Delete(context.Background(), localDir))
	assert.NoDirExists(t, localDir)

	p2pDir := filepath.Join(allowedDir, "p2p-dir")
	p2pChild := filepath.Join(p2pDir, "child.txt")
	require.NoError(t, os.MkdirAll(p2pDir, 0755))
	require.NoError(t, os.WriteFile(p2pChild, []byte("child"), 0644))
	err := tool.Delete(ctxkeys.WithP2PRequest(context.Background()), p2pDir)
	require.Error(t, err)
	assert.ErrorContains(t, err, "delete (p2p restricted)")
	assert.FileExists(t, p2pChild)

	targetFile := filepath.Join(allowedDir, "target.txt")
	linkPath := filepath.Join(allowedDir, "target-link.txt")
	require.NoError(t, os.WriteFile(targetFile, []byte("target"), 0644))
	requireSymlink(t, targetFile, linkPath)
	require.NoError(t, tool.Delete(context.Background(), linkPath))
	assert.NoFileExists(t, linkPath)
	assert.FileExists(t, targetFile)
}

func TestPathAccessAllowsSymlinkInsideAllowedAndBlocksSymlinkEscape(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	allowedDir := filepath.Join(tmpDir, "allowed")
	outsideDir := filepath.Join(tmpDir, "outside")
	require.NoError(t, os.MkdirAll(allowedDir, 0755))
	require.NoError(t, os.MkdirAll(outsideDir, 0755))

	insideTarget := filepath.Join(allowedDir, "inside.txt")
	insideLink := filepath.Join(allowedDir, "inside-link.txt")
	require.NoError(t, os.WriteFile(insideTarget, []byte("inside"), 0644))
	requireSymlink(t, insideTarget, insideLink)

	outsideTarget := filepath.Join(outsideDir, "outside.txt")
	escapeLink := filepath.Join(allowedDir, "escape-link.txt")
	require.NoError(t, os.WriteFile(outsideTarget, []byte("outside"), 0644))
	requireSymlink(t, outsideTarget, escapeLink)

	tool := New(Config{AllowedPaths: []string{allowedDir}})

	got, err := tool.Read(insideLink)
	require.NoError(t, err)
	assert.Equal(t, "inside", got)

	_, err = tool.Read(escapeLink)
	require.Error(t, err)
	assert.ErrorContains(t, err, "path not allowed")
}

func requireSymlink(t *testing.T, target, link string) {
	t.Helper()

	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink not supported or permitted in this environment: %v", err)
	}
}

func TestBlockedPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	blockedDir := filepath.Join(tmpDir, "secrets")
	allowedDir := filepath.Join(tmpDir, "public")

	require.NoError(t, os.MkdirAll(blockedDir, 0755))
	require.NoError(t, os.MkdirAll(allowedDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(blockedDir, "key.pem"), []byte("private"), 0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(allowedDir, "readme.txt"), []byte("hello"), 0644,
	))

	tests := []struct {
		give         string
		giveBlocked  []string
		wantErr      bool
		wantContains string
	}{
		{
			give:         filepath.Join(blockedDir, "key.pem"),
			giveBlocked:  []string{blockedDir},
			wantErr:      true,
			wantContains: "access denied: protected path",
		},
		{
			give:        filepath.Join(allowedDir, "readme.txt"),
			giveBlocked: []string{blockedDir},
			wantErr:     false,
		},
		{
			give:        filepath.Join(blockedDir, "key.pem"),
			giveBlocked: nil,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			tool := New(Config{BlockedPaths: tt.giveBlocked})
			_, err := tool.validatePath(tt.give)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.wantContains))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
