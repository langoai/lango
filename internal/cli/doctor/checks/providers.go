package checks

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/langoai/lango/internal/config"
)

// ProvidersCheck validates the AI provider configurations.
type ProvidersCheck struct{}

// Name returns the check name.
func (c *ProvidersCheck) Name() string {
	return "AI Providers"
}

// Run checks if providers are correctly configured.
func (c *ProvidersCheck) Run(ctx context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Configuration not loaded",
		}
	}

	var issues []string
	var passed []string

	// 1. Check modern providers map
	if len(cfg.Providers) > 0 {
		for id, pCfg := range cfg.Providers {
			if err := c.checkProvider(id, pCfg); err != nil {
				issues = append(issues, fmt.Sprintf("%s: %v", id, err))
			} else {
				passed = append(passed, id)
			}
		}
	}

	// 2. Check that agent.provider references a valid entry in the providers map
	if cfg.Agent.Provider != "" {
		if _, exists := cfg.Providers[cfg.Agent.Provider]; !exists && len(cfg.Providers) > 0 {
			issues = append(issues, fmt.Sprintf("agent.provider %q not found in providers map", cfg.Agent.Provider))
		}
	}

	// 3. Fallback: Check GOOGLE_API_KEY if absolutely nothing is configured
	if len(cfg.Providers) == 0 && cfg.Agent.Provider == "" {
		if os.Getenv("GOOGLE_API_KEY") != "" {
			return Result{
				Name:    c.Name(),
				Status:  StatusWarn,
				Message: "Implicit Gemini config found via GOOGLE_API_KEY",
				Details: "Please run 'lango onboard' to configure properly",
			}
		}
		return Result{
			Name:      c.Name(),
			Status:    StatusFail,
			Message:   "No AI providers configured",
			Details:   "Run 'lango onboard' to set up a provider",
			Fixable:   true,
			FixAction: "Run onboarding wizard",
		}
	}

	if len(issues) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("Provider issues found: %s", strings.Join(issues, ", ")),
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("Configured providers: %s", strings.Join(passed, ", ")),
	}
}

func (c *ProvidersCheck) checkProvider(name string, pCfg config.ProviderConfig) error {
	if pCfg.APIKey == "" {
		return fmt.Errorf("missing API key")
	}

	// Check for environment variable reference
	if len(pCfg.APIKey) > 3 && strings.HasPrefix(pCfg.APIKey, "${") && strings.HasSuffix(pCfg.APIKey, "}") {
		envVar := pCfg.APIKey[2 : len(pCfg.APIKey)-1]
		if os.Getenv(envVar) == "" {
			return fmt.Errorf("environment variable %s is not set", envVar)
		}
	}

	return nil
}

// Fix suggests running onboard.
func (c *ProvidersCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return Result{
		Name:    c.Name(),
		Status:  StatusSkip,
		Message: "Run 'lango onboard' to configure providers",
	}
}
