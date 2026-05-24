package checks

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
)

const minBaselineRecords = 3

// BootstrapTimingCheck compares current bootstrap phase timing against
// historical baseline from the diagnostics JSONL file.
type BootstrapTimingCheck struct{}

var _ BootstrapAwareCheck = (*BootstrapTimingCheck)(nil)

func (c *BootstrapTimingCheck) Name() string { return "Bootstrap Timing" }

func (c *BootstrapTimingCheck) Run(_ context.Context, _ *config.Config) Result {
	return Result{
		Name:    c.Name(),
		Status:  StatusSkip,
		Message: "Requires bootstrap context (run without --skip-bootstrap)",
	}
}

func (c *BootstrapTimingCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}

func (c *BootstrapTimingCheck) RunWithBootstrap(_ context.Context, _ *config.Config, boot *bootstrap.Result) Result {
	if boot == nil || len(boot.PhaseTiming) == 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "No phase timing data in current bootstrap",
		}
	}

	baseline, err := bootstrap.ReadTimingLog()
	if err != nil || len(baseline) == 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "No baseline timing file found — run doctor a few more times to build baseline",
		}
	}

	// Exclude the last entry: Pipeline.Execute appends the current run's timing
	// before doctor checks run, so the baseline would include the run being evaluated.
	// Single-writer assumption: only one bootstrap process appends at a time.
	if len(baseline) > 0 {
		baseline = baseline[:len(baseline)-1]
	}

	if len(baseline) < minBaselineRecords {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: fmt.Sprintf("Insufficient baseline (%d/%d records) — need %d runs for comparison", len(baseline), minBaselineRecords, minBaselineRecords),
		}
	}

	medians := computeMedians(baseline)

	var regressed []string
	var details []string
	for _, current := range boot.PhaseTiming {
		currentMs := current.Duration.Milliseconds()
		med, ok := medians[current.Phase]
		if !ok || med == 0 {
			details = append(details, fmt.Sprintf("  %s: %dms (no baseline)", current.Phase, currentMs))
			continue
		}
		ratio := float64(currentMs) / float64(med)
		details = append(details, fmt.Sprintf("  %s: %dms (baseline p50=%dms, %.1fx)", current.Phase, currentMs, med, ratio))
		if ratio > 2.0 {
			regressed = append(regressed, current.Phase)
		}
	}

	detailStr := strings.Join(details, "\n")

	if len(regressed) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: fmt.Sprintf("Bootstrap regression detected in %d phase(s): %s", len(regressed), strings.Join(regressed, ", ")),
			Details: detailStr,
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("Bootstrap timing within baseline (%d phases, %d baseline records)", len(boot.PhaseTiming), len(baseline)),
		Details: detailStr,
	}
}

func computeMedians(entries []bootstrap.TimingLogEntry) map[string]int64 {
	byPhase := make(map[string][]int64)
	for _, entry := range entries {
		for _, p := range entry.Phases {
			byPhase[p.Name] = append(byPhase[p.Name], p.DurationMs)
		}
	}

	medians := make(map[string]int64, len(byPhase))
	for name, durations := range byPhase {
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		mid := len(durations) / 2
		if len(durations)%2 == 0 && len(durations) > 1 {
			medians[name] = (durations[mid-1] + durations[mid]) / 2
		} else {
			medians[name] = durations[mid]
		}
	}
	return medians
}
