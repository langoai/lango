package app

import (
	"fmt"
	"math/big"
	"time"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/team"
)

// wireTeamBudgetBridge subscribes to team events and auto-manages budget lifecycle.
func wireTeamBudgetBridge(bus *eventbus.Bus, budgetEngine *budget.Engine, coord *team.Coordinator, log *zap.SugaredLogger) {
	// TeamFormed → allocate budget if team has budget > 0.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamFormedEvent) {
		t, err := coord.GetTeam(ev.TeamID)
		if err != nil {
			log.Debugw("team-budget bridge: team not found", "teamID", ev.TeamID, "error", err)
			return
		}

		if t.Budget <= 0 {
			return
		}

		totalBudget := floatToBudgetAmount(t.Budget)
		_, err = budgetEngine.Allocate(ev.TeamID, totalBudget)
		if err != nil {
			log.Warnw("team-budget bridge: allocate budget", "teamID", ev.TeamID, "error", err)
			return
		}

		log.Infow("team-budget bridge: budget allocated",
			"teamID", ev.TeamID,
			"budget", totalBudget.String(),
		)

		// Publish payment agreed event for each worker.
		members := t.Members()
		for _, m := range members {
			if m.Role == team.RoleWorker {
				bus.Publish(eventbus.TeamPaymentAgreedEvent{
					TeamID:    ev.TeamID,
					MemberDID: m.DID,
					Mode:      "prepay",
					Price:     totalBudget.String(),
				})
			}
		}
	})

	// TeamTaskDelegated → reserve estimated cost from budget.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamTaskDelegatedEvent) {
		// Estimate cost: base cost per worker invocation.
		estimatedCost := big.NewInt(int64(ev.Workers) * 100_000) // 0.1 USDC per worker
		releaseFn, err := budgetEngine.Reserve(ev.TeamID, estimatedCost)
		if err != nil {
			log.Debugw("team-budget bridge: reserve budget (may not be allocated)",
				"teamID", ev.TeamID, "error", err)
			return
		}

		// Release reservation after a timeout (will be committed by Record on completion).
		go func() {
			timer := time.NewTimer(5 * time.Minute)
			defer timer.Stop()
			<-timer.C
			releaseFn()
		}()

		log.Debugw("team-budget bridge: budget reserved",
			"teamID", ev.TeamID,
			"estimated", estimatedCost.String(),
		)
	})

	// TeamTaskCompleted → record actual cost.
	eventbus.SubscribeTyped(bus, func(ev eventbus.TeamTaskCompletedEvent) {
		// Calculate cost based on successful invocations.
		costPerInvocation := big.NewInt(100_000) // 0.1 USDC per invocation
		totalCost := new(big.Int).Mul(costPerInvocation, big.NewInt(int64(ev.Successful)))
		if totalCost.Sign() <= 0 {
			return
		}

		err := budgetEngine.Record(ev.TeamID, budget.SpendEntry{
			Amount:    totalCost,
			ToolName:  ev.ToolName,
			Reason:    fmt.Sprintf("team task: %d successful, %d failed", ev.Successful, ev.Failed),
			Timestamp: time.Now(),
		})
		if err != nil {
			log.Debugw("team-budget bridge: record spend (may not be allocated)",
				"teamID", ev.TeamID, "error", err)
			return
		}

		log.Debugw("team-budget bridge: spend recorded",
			"teamID", ev.TeamID,
			"amount", totalCost.String(),
			"tool", ev.ToolName,
		)
	})

	log.Info("team-budget bridge wired")
}

// floatToBudgetAmount converts a float64 dollar amount to budget wei (6 decimals).
// Note: this is identical to floatToUSDC but kept separate to avoid coupling.
func floatToBudgetAmount(amount float64) *big.Int {
	microUSDC := int64(amount * 1_000_000)
	return big.NewInt(microUSDC)
}
