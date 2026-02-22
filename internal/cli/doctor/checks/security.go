package checks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/session"
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
			// Check for known SQLite/SQLCipher errors related to encryption
			// "out of memory" (14) is often returned when opening an encrypted DB without a key
			// or with an incorrect page size/key.
			errMsg := err.Error()
			if strings.Contains(errMsg, "out of memory") || strings.Contains(errMsg, "file is not a database") {
				return Result{
					Name:    c.Name(),
					Status:  StatusWarn,
					Message: "Session database is encrypted and locked",
					Details: "The doctor command cannot verify the session database contents because it is encrypted.\n" +
						"Ensure your passphrase keyfile (~/.lango/keyfile) exists, or run in an interactive terminal.",
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

	message := "Security configuration verified"
	if len(issues) > 0 {
		message = fmt.Sprintf("Security issues found: %v", issues)
		// Format nicer if multiple?
		// "Security issues found:\n- Issue 1\n- Issue 2"
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
