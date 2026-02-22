package checks

import (
	"context"
	"os"
	"path/filepath"

	"github.com/langoai/lango/internal/config"
)

// DatabaseCheck validates the session database accessibility.
type DatabaseCheck struct{}

// Name returns the check name.
func (c *DatabaseCheck) Name() string {
	return "Session Database"
}

// Run checks if the session database is accessible.
func (c *DatabaseCheck) Run(ctx context.Context, cfg *config.Config) Result {
	dbPath := c.resolveDatabasePath(cfg)

	// Check if directory exists
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return Result{
			Name:      c.Name(),
			Status:    StatusFail,
			Message:   "Database directory does not exist",
			Details:   dir,
			Fixable:   true,
			FixAction: "Create database directory",
		}
	}

	// Check if directory is writable
	testFile := filepath.Join(dir, ".lango-test-write")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Database directory not writable",
			Details: err.Error(),
		}
	}
	os.Remove(testFile)

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Database file does not exist (will be created on first run)",
			Details: dbPath,
		}
	}

	// Try to open the file to verify it's readable
	file, err := os.OpenFile(dbPath, os.O_RDWR, 0644)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Cannot open database file",
			Details: err.Error(),
		}
	}
	file.Close()

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Session database accessible",
		Details: dbPath,
	}
}

// Fix creates the database directory if missing.
func (c *DatabaseCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	dbPath := c.resolveDatabasePath(cfg)
	dir := filepath.Dir(dbPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Failed to create database directory",
			Details: err.Error(),
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Created database directory",
		Details: dir,
	}
}

// resolveDatabasePath determines the database path from config or defaults.
func (c *DatabaseCheck) resolveDatabasePath(cfg *config.Config) string {
	if cfg != nil && cfg.Session.DatabasePath != "" {
		path := cfg.Session.DatabasePath
		// Expand ~ to home directory
		if len(path) > 0 && path[0] == '~' {
			home, err := os.UserHomeDir()
			if err == nil {
				path = filepath.Join(home, path[1:])
			}
		}
		return path
	}

	// Default path
	home, err := os.UserHomeDir()
	if err != nil {
		return "sessions.db"
	}
	return filepath.Join(home, ".lango", "sessions.db")
}
