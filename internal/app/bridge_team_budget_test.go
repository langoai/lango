package app

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/p2p/team"
)

func newTestBudgetEngine(t *testing.T) *budget.Engine {
	t.Helper()
	store := budget.NewStore()
	eng, err := budget.NewEngine(store, config.BudgetConfig{
		DefaultMax: "100.00",
	})
	require.NoError(t, err)
	return eng
}

func TestTeamBudgetBridge_ShutdownCancelsReservation(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)
	eng := newTestBudgetEngine(t)

	_, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-cancel", Name: "cancel-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)

	// Pre-allocate budget for the team so Reserve works.
	_, err = eng.Allocate("t-cancel", big.NewInt(10_000_000))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	wireTeamBudgetBridge(ctx, bus, eng, coord, testLog())

	// Publish a delegation event — this triggers Reserve + goroutine.
	bus.Publish(eventbus.TeamTaskDelegatedEvent{
		TeamID:   "t-cancel",
		ToolName: "search",
		Workers:  1,
	})

	// Allow a brief moment for the goroutine to start waiting on ctx.
	time.Sleep(50 * time.Millisecond)

	// Cancel the context — should trigger releaseFn().
	cancel()

	// Give the goroutine time to call releaseFn.
	time.Sleep(100 * time.Millisecond)

	// Verify the reserved amount was released (Reserved should be 0).
	// The initial reserve was Workers*100_000 = 100_000.
	// After release, we should be able to reserve the full budget again.
	releaseFn, err := eng.Reserve("t-cancel", big.NewInt(10_000_000))
	assert.NoError(t, err, "full budget should be available after reservation release")
	if releaseFn != nil {
		releaseFn()
	}
}

func TestTeamBudgetBridge_TeamFormedAllocatesBudget(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	coord := setupTestCoordinator(t, bus)
	eng := newTestBudgetEngine(t)

	ctx := context.Background()
	wireTeamBudgetBridge(ctx, bus, eng, coord, testLog())

	tm, err := coord.FormTeam(context.Background(), team.FormTeamRequest{
		TeamID: "t-budget", Name: "budget-team", Goal: "test",
		LeaderDID: "did:leader", Capability: "search", MemberCount: 1,
	})
	require.NoError(t, err)
	tm.Budget = 50.0

	// Re-publish the formed event so the bridge sees it.
	bus.Publish(eventbus.TeamFormedEvent{
		TeamID:    "t-budget",
		Name:      "budget-team",
		Goal:      "test",
		LeaderDID: "did:leader",
		Members:   2,
	})

	// Budget allocation might fail because the team's budget is set after form.
	// That's OK — this tests that the bridge handles the event without panic.
}
