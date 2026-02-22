package checks

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// ObservationalMemoryCheck validates observational memory configuration.
type ObservationalMemoryCheck struct{}

// Name returns the check name.
func (c *ObservationalMemoryCheck) Name() string {
	return "Observational Memory"
}

// Run checks observational memory configuration validity.
func (c *ObservationalMemoryCheck) Run(ctx context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	om := cfg.ObservationalMemory

	if !om.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Observational memory is disabled",
		}
	}

	var issues []string
	status := StatusPass

	// Check threshold values are positive
	if om.MessageTokenThreshold <= 0 {
		issues = append(issues, fmt.Sprintf(
			"messageTokenThreshold must be positive (got %d)", om.MessageTokenThreshold))
		status = StatusFail
	}

	if om.ObservationTokenThreshold <= 0 {
		issues = append(issues, fmt.Sprintf(
			"observationTokenThreshold must be positive (got %d)", om.ObservationTokenThreshold))
		status = StatusFail
	}

	if om.MaxMessageTokenBudget <= 0 {
		issues = append(issues, fmt.Sprintf(
			"maxMessageTokenBudget must be positive (got %d)", om.MaxMessageTokenBudget))
		status = StatusFail
	}

	// Check budget > threshold consistency
	if om.MaxMessageTokenBudget > 0 && om.MessageTokenThreshold > 0 &&
		om.MaxMessageTokenBudget <= om.MessageTokenThreshold {
		issues = append(issues,
			"maxMessageTokenBudget should be greater than messageTokenThreshold")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	// Check custom provider exists in providers map
	if om.Provider != "" && cfg.Providers != nil {
		if _, ok := cfg.Providers[om.Provider]; !ok {
			issues = append(issues, fmt.Sprintf(
				"provider '%s' not found in providers configuration", om.Provider))
			if status < StatusWarn {
				status = StatusWarn
			}
		}
	}

	if len(issues) == 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "Observational memory configuration valid",
		}
	}

	message := "Observational memory issues:\n"
	for _, issue := range issues {
		message += fmt.Sprintf("- %s\n", issue)
	}

	return Result{
		Name:    c.Name(),
		Status:  status,
		Message: message,
	}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *ObservationalMemoryCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
