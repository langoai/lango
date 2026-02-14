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
	tool := New(Config{})
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write
	content := "hello\nworld"
	if err := tool.Write(testFile, content); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Read
	result, err := tool.Read(testFile)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if result != content {
		t.Errorf("expected %q, got %q", content, result)
	}
}

func TestReadLines(t *testing.T) {
	tool := New(Config{})
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "lines.txt")

	content := "line1\nline2\nline3\nline4\nline5"
	if err := tool.Write(testFile, content); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	result, err := tool.ReadLines(testFile, 2, 4)
	if err != nil {
		t.Fatalf("readLines failed: %v", err)
	}

	expected := "line2\nline3\nline4"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEdit(t *testing.T) {
	tool := New(Config{})
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "edit.txt")

	content := "line1\nold\nline3"
	if err := tool.Write(testFile, content); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if err := tool.Edit(testFile, 2, 2, "new"); err != nil {
		t.Fatalf("edit failed: %v", err)
	}

	result, _ := tool.Read(testFile)
	expected := "line1\nnew\nline3"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestListDir(t *testing.T) {
	tool := New(Config{})
	tmpDir := t.TempDir()

	// Create some files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("b"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	files, err := tool.ListDir(tmpDir)
	if err != nil {
		t.Fatalf("listDir failed: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("expected 3 entries, got %d", len(files))
	}
}

func TestPathValidation(t *testing.T) {
	tool := New(Config{
		AllowedPaths: []string{"/tmp/allowed"},
	})

	// Should fail for paths outside allowed
	_, err := tool.validatePath("/etc/passwd")
	if err == nil {
		t.Error("expected error for disallowed path")
	}
}

func TestFileSizeLimit(t *testing.T) {
	tool := New(Config{MaxReadSize: 10})
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Write file larger than limit
	os.WriteFile(testFile, []byte("this is larger than 10 bytes"), 0644)

	_, err := tool.Read(testFile)
	if err == nil {
		t.Error("expected error for large file")
	}
}

func TestBlockedPaths(t *testing.T) {
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
