package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
		give       string
		giveOffset int
		giveLimit  int
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
