package checks

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// RunLedgerCheck validates RunLedger configuration invariants.
type RunLedgerCheck struct{}

// Name returns the check name.
func (c *RunLedgerCheck) Name() string {
	return "RunLedger"
}

// Run checks RunLedger configuration validity.
func (c *RunLedgerCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	rl := cfg.RunLedger
	if !rl.Enabled {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "RunLedger is disabled"}
	}

	if rl.StaleTTL <= 0 {
		return Result{Name: c.Name(), Status: StatusFail, Message: "runLedger.staleTtl must be greater than 0"}
	}
	if rl.ValidatorTimeout <= 0 {
		return Result{Name: c.Name(), Status: StatusFail, Message: "runLedger.validatorTimeout must be greater than 0"}
	}
	if rl.MaxRunHistory < 0 {
		return Result{Name: c.Name(), Status: StatusFail, Message: "runLedger.maxRunHistory must be 0 or greater"}
	}
	if rl.PlannerMaxRetries < 0 {
		return Result{Name: c.Name(), Status: StatusFail, Message: "runLedger.plannerMaxRetries must be 0 or greater"}
	}
	if rl.AuthoritativeRead && !rl.WriteThrough {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "runLedger.authoritativeRead requires runLedger.writeThrough",
			Details: "Enable writeThrough before authoritativeRead so ledger snapshots stay fed by write paths.",
		}
	}

	return Result{
		Name:   c.Name(),
		Status: StatusPass,
		Message: fmt.Sprintf(
			"RunLedger configured (writeThrough=%t, authoritativeRead=%t, workspaceIsolation=%t)",
			rl.WriteThrough,
			rl.AuthoritativeRead,
			rl.WorkspaceIsolation,
		),
	}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *RunLedgerCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
