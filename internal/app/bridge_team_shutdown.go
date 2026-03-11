package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/team"
)

// initTeamShutdownBridge wires budget events to team graceful shutdown.
// BudgetExhaustedEvent triggers GracefulShutdown for the corresponding team.
// BudgetAlertEvent (80% threshold) triggers a TeamBudgetWarningEvent.
func initTeamShutdownBridge(bus *eventbus.Bus, coordinator *team.Coordinator, log *zap.SugaredLogger) {
	// BudgetAlertEvent → TeamBudgetWarningEvent when threshold >= 0.8.
	eventbus.SubscribeTyped(bus, func(ev eventbus.BudgetAlertEvent) {
		if ev.Threshold < 0.8 {
			return
		}

		t, err := coordinator.GetTeam(ev.TaskID)
		if err != nil {
			log.Debugw("team-shutdown bridge: team not found for budget alert",
				"taskID", ev.TaskID, "error", err)
			return
		}

		bus.Publish(eventbus.TeamBudgetWarningEvent{
			TeamID:    ev.TaskID,
			Threshold: ev.Threshold,
			Spent:     t.Spent,
			Budget:    t.Budget,
		})

		log.Infow("team-shutdown bridge: budget warning published",
			"teamID", ev.TaskID,
			"threshold", ev.Threshold,
			"spent", t.Spent,
			"budget", t.Budget,
		)
	})

	// BudgetExhaustedEvent → GracefulShutdown.
	eventbus.SubscribeTyped(bus, func(ev eventbus.BudgetExhaustedEvent) {
		log.Infow("team-shutdown bridge: budget exhausted, initiating graceful shutdown",
			"taskID", ev.TaskID)

		if err := coordinator.GracefulShutdown(context.Background(), ev.TaskID, "budget exhausted"); err != nil {
			log.Debugw("team-shutdown bridge: graceful shutdown",
				"taskID", ev.TaskID, "error", err)
		}
	})

	log.Info("team-shutdown bridge wired")
}
