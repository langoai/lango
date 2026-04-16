package app

import (
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
)

// TeamMetrics holds in-memory counters for team task delegation analysis.
type TeamMetrics struct {
	Delegations    atomic.Int64
	TotalWorkers   atomic.Int64
	TotalSuccesses atomic.Int64
	TotalFailures  atomic.Int64
}

// wireTeamMetricsBridge subscribes to team task events and logs structured
// metrics for P1 analysis (shared task coordination feasibility).
func wireTeamMetricsBridge(bus *eventbus.Bus, metrics *TeamMetrics, log *zap.SugaredLogger) {
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamTaskDelegatedEvent) {
		metrics.Delegations.Add(1)
		metrics.TotalWorkers.Add(int64(ev.Workers))
		log.Infow("team-metrics: task delegated",
			"teamID", ev.TeamID,
			"tool", ev.ToolName,
			"workers", ev.Workers,
			"totalDelegations", metrics.Delegations.Load(),
		)
	})

	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamTaskCompletedEvent) {
		metrics.TotalSuccesses.Add(int64(ev.Successful))
		metrics.TotalFailures.Add(int64(ev.Failed))

		total := ev.Successful + ev.Failed
		dupeRatio := 0.0
		if total > 1 && ev.Successful > 0 {
			dupeRatio = float64(ev.Successful-1) / float64(total)
		}

		log.Infow("team-metrics: task completed",
			"teamID", ev.TeamID,
			"tool", ev.ToolName,
			"successful", ev.Successful,
			"failed", ev.Failed,
			"avgDuration", ev.Duration.String(),
			"dupeRatioEstimate", dupeRatio,
			"lifetimeSuccesses", metrics.TotalSuccesses.Load(),
			"lifetimeFailures", metrics.TotalFailures.Load(),
		)
	})

	log.Info("team-metrics bridge wired")
}
