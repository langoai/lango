package checks

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/config"
)

// APIKeySecurityCheck validates that API keys use environment variable references
// rather than plaintext values.
type APIKeySecurityCheck struct{}

// Name returns the check name.
func (c *APIKeySecurityCheck) Name() string {
	return "API Key Security"
}

// Run checks if any provider API keys are stored as plaintext.
func (c *APIKeySecurityCheck) Run(ctx context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Configuration not loaded",
		}
	}

	if len(cfg.Providers) == 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "No providers configured",
		}
	}

	var plaintext []string
	total := 0

	for id, pCfg := range cfg.Providers {
		if pCfg.APIKey == "" {
			continue
		}
		total++
		if !isEnvVarReference(pCfg.APIKey) {
			plaintext = append(plaintext, id)
		}
	}

	if total == 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "No API keys configured",
		}
	}

	if len(plaintext) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: fmt.Sprintf("Inline API keys for: %s", strings.Join(plaintext, ", ")),
			Details: "Keys stored directly in an encrypted profile are safe. For portability across environments, consider using ${ENV_VAR} references instead.",
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "All API keys secured",
	}
}

// Fix suggests using environment variables.
func (c *APIKeySecurityCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return Result{
		Name:    c.Name(),
		Status:  StatusSkip,
		Message: "Replace plaintext API keys with ${ENV_VAR} references in your configuration",
	}
}

// isEnvVarReference checks if a string is an environment variable reference like ${VAR_NAME}.
func isEnvVarReference(s string) bool {
	return strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}")
}
