package turnrunner

import (
	"context"

	"google.golang.org/adk/runner"

	"github.com/langoai/lango/internal/live"
)

// RunnerLiveExecutor is the default live.LiveExecutor backed by an
// *adk.runner.Runner. The LiveExecutor interface itself is declared in
// internal/live (see live.LiveExecutor) to avoid an import cycle —
// turnrunner imports live, so live cannot reference turnrunner.
//
// Phase 1: cockpit voice mode is the only consumer; text-only mode.
// Phase 2: audio I/O will be wired through live.Config.AudioSink/Source.
type RunnerLiveExecutor struct {
	Runner *runner.Runner
}

// StartLive implements live.LiveExecutor.
func (e *RunnerLiveExecutor) StartLive(ctx context.Context, userID, sessionID string, cfg live.Config) (*live.Session, error) {
	return live.New(ctx, e.Runner, userID, sessionID, cfg)
}

// Compile-time assertion that RunnerLiveExecutor satisfies live.LiveExecutor.
var _ live.LiveExecutor = (*RunnerLiveExecutor)(nil)
