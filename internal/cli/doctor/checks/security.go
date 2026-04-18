package checks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/sqlitedriver"
)

// SecurityCheck checks security configuration and state
type SecurityCheck struct{}

func (c *SecurityCheck) Name() string {
	return "Security Configuration"
}

func (c *SecurityCheck) Run(ctx context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	var issues []string
	status := StatusPass

	// 1. Check Provider
	switch cfg.Security.Signer.Provider {
	case "enclave":
		// Most secure option — no warnings
	case "rpc":
		// Production-ready option — no warnings
	case "local":
		issues = append(issues, "Using 'local' security provider (not recommended for production)")
		if status < StatusWarn {
			status = StatusWarn
		}
	default:
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("Unknown security provider: %s", cfg.Security.Signer.Provider),
		}
	}

	// 2. Check Database State (Salt/Checksum)
	if cfg.Session.DatabasePath == "" {
		issues = append(issues, "No session database path configured")
		status = StatusWarn
	} else {
		store, err := session.NewEntStore(cfg.Session.DatabasePath)
		if err != nil {
			if errors.Is(err, sqlitedriver.ErrLegacyEncryptedOrUnreadableDB) {
				return Result{
					Name:    c.Name(),
					Status:  StatusWarn,
					Message: "Session database is a legacy encrypted or unreadable DB",
					Details: "This runtime no longer supports SQLCipher database files.\n" +
						"Use an older build to export or decrypt the database before upgrading.",
				}
			}
			return Result{
				Name:    c.Name(),
				Status:  StatusFail,
				Message: fmt.Sprintf("Failed to access session store: %v", err),
			}
		}
		defer store.Close()

		// Check Salt equality/existence logic?
		// Just check if checksum exists.
		_, err = store.GetChecksum("default")
		if err != nil {
			// If salt missing, GetChecksum might fail or return nil?
			// EntStore implementation: GetChecksum returns error if query fails.
			// Salt/Checksum usually exist together.
			issues = append(issues, "Passphrase checksum not found in database (run 'lango security migrate-passphrase'?)")
			if status < StatusWarn {
				status = StatusWarn
			}
		}
	}

	// 3. Check DB encryption status.
	if cfg.Security.DBEncryption.Enabled {
		issues = append(issues, "security.dbEncryption is deprecated and ignored by this runtime")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	dbPath := cfg.Session.DatabasePath
	if strings.HasPrefix(dbPath, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			dbPath = filepath.Join(h, dbPath[2:])
		}
	}
	if bootstrap.IsDBEncrypted(dbPath) {
		issues = append(issues, "Session database header indicates a legacy encrypted or unreadable DB; export it with an older build before upgrading")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	message := "Security configuration verified"
	if len(issues) > 0 {
		message = "Security checks returned warnings:\n"
		for _, issue := range issues {
			message += fmt.Sprintf("- %s\n", issue)
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  status,
		Message: message,
	}
}

func (c *SecurityCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}

// CompanionConnectionCheck checks if the Gateway is running and if companions are connected.
type CompanionConnectionCheck struct{}

func (c *CompanionConnectionCheck) Name() string {
	return "Companion Connectivity"
}

func (c *CompanionConnectionCheck) Run(ctx context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Server.WebSocketEnabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "WebSockets disabled in config",
		}
	}

	// Check if server is reachable
	statusURL := fmt.Sprintf("http://localhost:%d/status", cfg.Server.Port)
	if cfg.Server.Host != "" && cfg.Server.Host != "0.0.0.0" {
		statusURL = fmt.Sprintf("http://%s:%d/status", cfg.Server.Host, cfg.Server.Port)
	}

	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(statusURL)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Gateway server not reachable",
			Details: fmt.Sprintf("Could not connect to %s. Ensure the server is running.\nError: %v", statusURL, err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: fmt.Sprintf("Gateway returned status %d", resp.StatusCode),
		}
	}

	var status struct {
		Clients int `json:"clients"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Invalid status response from gateway",
		}
	}

	if status.Clients > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: fmt.Sprintf("Gateway reachable (%d clients connected)", status.Clients),
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass, // Pass but with detail note
		Message: "Gateway reachable (no companions connected)",
		Details: "Connect a companion app to enable security features.",
	}
}

func (c *CompanionConnectionCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
