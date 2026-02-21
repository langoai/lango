package filesystem

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/langowarny/lango/internal/logging"
)

var logger = logging.SubsystemSugar("tool.filesystem")

// Config holds filesystem tool configuration
type Config struct {
	MaxReadSize  int64    // maximum file size to read
	AllowedPaths []string // allowed base paths (empty = all)
	BlockedPaths []string // paths that are always denied
}

// Tool provides filesystem operations
type Tool struct {
	config Config
}

// FileInfo represents file metadata
type FileInfo struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"isDir"`
	ModTime int64  `json:"modTime"`
	Mode    string `json:"mode"`
}

// New creates a new filesystem tool
func New(cfg Config) *Tool {
	if cfg.MaxReadSize == 0 {
		cfg.MaxReadSize = 10 * 1024 * 1024 // 10MB default
	}
	return &Tool{config: cfg}
}

// Read reads a file and returns its contents
func (t *Tool) Read(path string) (string, error) {
	absPath, err := t.validatePath(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("file not found: %s", path)
	}

	if info.IsDir() {
		return "", fmt.Errorf("cannot read directory: %s", path)
	}

	if info.Size() > t.config.MaxReadSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), t.config.MaxReadSize)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	logger.Infow("file read", "path", absPath, "size", len(content))
	return string(content), nil
}

// ReadLines reads specific lines from a file
func (t *Tool) ReadLines(path string, startLine, endLine int) (string, error) {
	absPath, err := t.validatePath(path)
	if err != nil {
		return "", err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && lineNum <= endLine {
			lines = append(lines, scanner.Text())
		}
		if lineNum > endLine {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}

// Write writes content to a file (atomic write)
func (t *Tool) Write(path, content string) error {
	absPath, err := t.validatePath(path)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Write to temp file first (atomic write)
	tempPath := absPath + ".tmp"
	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Rename to final path
	if err := os.Rename(tempPath, absPath); err != nil {
		os.Remove(tempPath) // cleanup
		return fmt.Errorf("rename file: %w", err)
	}

	logger.Infow("file written", "path", absPath, "size", len(content))
	return nil
}

// Edit replaces content in a file between specified lines
func (t *Tool) Edit(path string, startLine, endLine int, newContent string) error {
	absPath, err := t.validatePath(path)
	if err != nil {
		return err
	}

	// Read existing content
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	// Validate line range
	if startLine < 1 || startLine > len(lines)+1 {
		return fmt.Errorf("invalid start line: %d (file has %d lines)", startLine, len(lines))
	}
	if endLine < startLine {
		return fmt.Errorf("end line must be >= start line")
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}

	// Replace lines
	newLines := strings.Split(newContent, "\n")
	result := make([]string, 0, len(lines)-endLine+startLine-1+len(newLines))
	result = append(result, lines[:startLine-1]...)
	result = append(result, newLines...)
	if endLine < len(lines) {
		result = append(result, lines[endLine:]...)
	}

	// Write back
	return t.Write(path, strings.Join(result, "\n"))
}

// ListDir lists contents of a directory
func (t *Tool) ListDir(path string) ([]FileInfo, error) {
	absPath, err := t.validatePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	result := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		result = append(result, FileInfo{
			Path:    filepath.Join(absPath, entry.Name()),
			Name:    entry.Name(),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime().Unix(),
			Mode:    info.Mode().String(),
		})
	}

	return result, nil
}

// Delete removes a file or directory
func (t *Tool) Delete(path string) error {
	absPath, err := t.validatePath(path)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	logger.Infow("file deleted", "path", absPath)
	return nil
}

// Exists checks if a path exists
func (t *Tool) Exists(path string) (bool, error) {
	absPath, err := t.validatePath(path)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// Mkdir creates a directory
func (t *Tool) Mkdir(path string) error {
	absPath, err := t.validatePath(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	logger.Infow("directory created", "path", absPath)
	return nil
}

// Copy copies a file
func (t *Tool) Copy(src, dst string) error {
	srcPath, err := t.validatePath(src)
	if err != nil {
		return err
	}

	dstPath, err := t.validatePath(dst)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	logger.Infow("file copied", "src", srcPath, "dst", dstPath)
	return nil
}

// validatePath checks if a path is safe and converts to absolute
func (t *Tool) validatePath(path string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Clean the path to prevent traversal
	absPath = filepath.Clean(absPath)

	// Check against blocked paths
	for _, blocked := range t.config.BlockedPaths {
		absBlocked, err := filepath.Abs(blocked)
		if err != nil {
			continue
		}
		absBlocked = filepath.Clean(absBlocked)
		if strings.HasPrefix(absPath, absBlocked) {
			return "", fmt.Errorf("access denied: protected path")
		}
	}

	// Check against allowed paths
	if len(t.config.AllowedPaths) > 0 {
		allowed := false
		for _, base := range t.config.AllowedPaths {
			absBase, _ := filepath.Abs(base)
			if strings.HasPrefix(absPath, absBase) {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("path not allowed: %s", path)
		}
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		// Re-check cleaned path is still within bounds
		cleanPath := filepath.Clean(path)
		if cleanPath != path && strings.Contains(cleanPath, "..") {
			return "", fmt.Errorf("path traversal not allowed: %s", path)
		}
	}

	return absPath, nil
}
