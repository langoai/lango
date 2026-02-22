package checks

import (
	"context"
	"os"
	"path/filepath"

	"github.com/langoai/lango/internal/config"
)

// ConfigCheck validates the encrypted configuration profile.
type ConfigCheck struct{}

// Name returns the check name.
func (c *ConfigCheck) Name() string {
	return "Configuration Profile"
}

// Run checks if an encrypted configuration profile exists and is valid.
func (c *ConfigCheck) Run(ctx context.Context, cfg *config.Config) Result {
	if cfg == nil {
		// No config loaded — check if the DB file exists to give a more specific message.
		dbPath := defaultDBPath()
		if _, err := os.Stat(dbPath); err != nil {
			return Result{
				Name:      c.Name(),
				Status:    StatusFail,
				Message:   "Encrypted profile database not found",
				Details:   "No lango.db found at " + dbPath + ". Run 'lango onboard' to set up.",
				Fixable:   true,
				FixAction: "Run 'lango onboard' to create an encrypted profile",
			}
		}
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "No active configuration profile loaded",
			Details: "The profile database exists but no active profile could be loaded. Run 'lango onboard' to configure.",
		}
	}

	if err := config.Validate(cfg); err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Configuration has validation warnings",
			Details: err.Error(),
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Encrypted configuration profile valid",
		Details: defaultDBPath(),
	}
}

// Fix guides the user to run 'lango onboard' for profile setup.
func (c *ConfigCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return Result{
		Name:      c.Name(),
		Status:    StatusFail,
		Message:   "Run 'lango onboard' to set up your configuration",
		Details:   "Interactive setup is required. Run: lango onboard",
		Fixable:   true,
		FixAction: "Run 'lango onboard'",
	}
}

// defaultDBPath returns the expected path to the encrypted profile database.
func defaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".lango", "lango.db")
}
