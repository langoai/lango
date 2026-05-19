package team

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/economy/budget"
	"github.com/langoai/lango/internal/economy/escrow"
)

type wave54Settler struct {
	lockErr    error
	releaseErr error
	locked     []*big.Int
	released   []*big.Int
}

func (s *wave54Settler) Lock(_ context.Context, _ string, amount *big.Int) error {
	if s.lockErr != nil {
		return s.lockErr
	}
	s.locked = append(s.locked, new(big.Int).Set(amount))
	return nil
}

func (s *wave54Settler) Release(_ context.Context, _ string, amount *big.Int) error {
	if s.releaseErr != nil {
		return s.releaseErr
	}
	s.released = append(s.released, new(big.Int).Set(amount))
	return nil
}

func (s *wave54Settler) Refund(context.Context, string, *big.Int) error {
	return nil
}

func TestWave54BuildEscrowToolsConstructsDangerousP2PTools(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)
	tools := BuildEscrowTools(coord, wave54EscrowEngine(&wave54Settler{}), nil)
	require.Len(t, tools, 2)

	formTool := findToolByName(tools, "team_form_with_budget")
	require.NotNil(t, formTool)
	assert.Equal(t, agent.SafetyLevelDangerous, formTool.SafetyLevel)
	assert.Equal(t, "p2p", formTool.Capability.Category)
	assert.Equal(t, agent.ActivityExecute, formTool.Capability.Activity)
	assert.NotNil(t, formTool.Handler)
	assert.ElementsMatch(t,
		[]string{"name", "goal", "capability", "memberCount", "leaderDid", "budget"},
		requiredParameters(t, formTool),
	)

	completeTool := findToolByName(tools, "team_complete_milestone")
	require.NotNil(t, completeTool)
	assert.Equal(t, agent.SafetyLevelDangerous, completeTool.SafetyLevel)
	assert.Equal(t, "p2p", completeTool.Capability.Category)
	assert.Equal(t, agent.ActivityExecute, completeTool.Capability.Activity)
	assert.NotNil(t, completeTool.Handler)
	assert.ElementsMatch(t, []string{"escrowId", "milestoneId"}, requiredParameters(t, completeTool))
}

func TestWave54TeamFormWithBudgetCreatesEscrowAndBudget(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)
	settler := &wave54Settler{}
	escrowEngine := wave54EscrowEngine(settler)
	budgetEngine, err := budget.NewEngine(budget.NewStore(), config.BudgetConfig{})
	require.NoError(t, err)
	tool := findToolByName(BuildEscrowTools(coord, escrowEngine, budgetEngine), "team_form_with_budget")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"name":        "budget-team",
		"goal":        "ship deterministic tests",
		"capability":  "search",
		"memberCount": float64(2),
		"leaderDid":   "did:leader",
		"budget":      float64(12.5),
		"milestones": []interface{}{
			map[string]interface{}{"description": "first", "amount": float64(5)},
			map[string]interface{}{"description": "second", "amount": float64(7.5)},
		},
	})
	require.NoError(t, err)
	result := requireEscrowToolMap(t, got)
	assert.NotEmpty(t, result["teamId"])
	assert.Equal(t, "budget-team", result["name"])
	assert.Equal(t, "ship deterministic tests", result["goal"])
	assert.Equal(t, string(StatusActive), result["status"])
	assert.NotEmpty(t, result["escrowId"])
	assert.Equal(t, result["teamId"], result["budgetId"])
	assert.Equal(t, float64(12.5), result["budget"])
	assert.Equal(t, 2, result["milestones"])
	assert.NotEmpty(t, result["createdAt"])
	require.Len(t, result["members"], 3)

	entry, err := escrowEngine.Get(result["escrowId"].(string))
	require.NoError(t, err)
	assert.Equal(t, "did:leader", entry.BuyerDID)
	assert.Contains(t, []string{"did:worker1", "did:worker2"}, entry.SellerDID)
	assert.Equal(t, "Team budget-team: ship deterministic tests", entry.Reason)
	assert.Equal(t, "12500000", entry.TotalAmount.String())
	require.Len(t, entry.Milestones, 2)
	assert.Equal(t, "first", entry.Milestones[0].Description)
	assert.Equal(t, "5000000", entry.Milestones[0].Amount.String())
	assert.Empty(t, settler.locked, "creating escrow should not fund on-chain state")
}

func TestWave54TeamFormWithBudgetAutoSplitsAcrossWorkers(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)
	escrowEngine := wave54EscrowEngine(&wave54Settler{})
	budgetEngine, err := budget.NewEngine(budget.NewStore(), config.BudgetConfig{})
	require.NoError(t, err)
	tool := findToolByName(BuildEscrowTools(coord, escrowEngine, budgetEngine), "team_form_with_budget")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"name":        "split-team",
		"goal":        "auto split",
		"capability":  "search",
		"memberCount": float64(2),
		"leaderDid":   "did:leader",
		"budget":      float64(10),
	})
	require.NoError(t, err)
	result := requireEscrowToolMap(t, got)
	assert.Equal(t, result["teamId"], result["budgetId"])

	entry, err := escrowEngine.Get(result["escrowId"].(string))
	require.NoError(t, err)
	require.Len(t, entry.Milestones, 2)
	assert.Equal(t, "5000000", entry.Milestones[0].Amount.String())
	assert.Equal(t, "5000000", entry.Milestones[1].Amount.String())
	assert.Contains(t, entry.Milestones[0].Description, "did:worker")
}

func TestWave54TeamFormWithBudgetWrapsTeamAndEscrowErrors(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)
	tool := findToolByName(BuildEscrowTools(coord, wave54EscrowEngine(&wave54Settler{}), nil), "team_form_with_budget")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"name":        "missing-workers",
		"goal":        "cannot select",
		"capability":  "nonexistent-capability",
		"memberCount": float64(1),
		"leaderDid":   "did:leader",
		"budget":      float64(1),
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "form team:")

	milestones := make([]interface{}, 11)
	for i := range milestones {
		milestones[i] = map[string]interface{}{"description": "too many", "amount": float64(1)}
	}
	got, err = tool.Handler(context.Background(), map[string]interface{}{
		"name":        "bad-milestones",
		"goal":        "escrow fails",
		"capability":  "search",
		"memberCount": float64(1),
		"leaderDid":   "did:leader",
		"budget":      float64(1),
		"milestones":  milestones,
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "create escrow:")
}

func TestWave54TeamCompleteMilestoneReleasesWhenAllMilestonesComplete(t *testing.T) {
	t.Parallel()

	settler := &wave54Settler{}
	escrowEngine := wave54EscrowEngine(settler)
	entry := wave54ActiveEscrow(t, escrowEngine)
	coord, _ := setupCoordinator(t)
	tool := findToolByName(BuildEscrowTools(coord, escrowEngine, nil), "team_complete_milestone")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"escrowId":    entry.ID,
		"milestoneId": entry.Milestones[0].ID,
	})
	require.NoError(t, err)
	result := requireEscrowToolMap(t, got)
	assert.Equal(t, entry.ID, result["escrowId"])
	assert.Equal(t, entry.Milestones[0].ID, result["milestoneId"])
	assert.Equal(t, string(escrow.StatusReleased), result["status"])
	assert.Equal(t, 1, result["completedMilestones"])
	assert.Equal(t, 1, result["totalMilestones"])
	assert.Equal(t, true, result["allCompleted"])
	assert.Equal(t, true, result["released"])
	require.Len(t, settler.released, 1)
	assert.Equal(t, "2500000", settler.released[0].String())

	updated, err := escrowEngine.Get(entry.ID)
	require.NoError(t, err)
	assert.Equal(t, "manual completion", updated.Milestones[0].Evidence)
}

func TestWave54TeamCompleteMilestoneReportsReleaseError(t *testing.T) {
	t.Parallel()

	settler := &wave54Settler{releaseErr: errors.New("release unavailable")}
	escrowEngine := wave54EscrowEngine(settler)
	entry := wave54ActiveEscrow(t, escrowEngine)
	coord, _ := setupCoordinator(t)
	tool := findToolByName(BuildEscrowTools(coord, escrowEngine, nil), "team_complete_milestone")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"escrowId":    entry.ID,
		"milestoneId": entry.Milestones[0].ID,
		"evidence":    "proof",
	})
	require.NoError(t, err)
	result := requireEscrowToolMap(t, got)
	assert.Equal(t, string(escrow.StatusCompleted), result["status"])
	assert.Equal(t, "release funds: release unavailable", result["releaseError"])
	assert.NotContains(t, result, "released")
}

func TestWave54TeamCompleteMilestoneWrapsEngineError(t *testing.T) {
	t.Parallel()

	coord, _ := setupCoordinator(t)
	tool := findToolByName(BuildEscrowTools(coord, wave54EscrowEngine(&wave54Settler{}), nil), "team_complete_milestone")
	require.NotNil(t, tool)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"escrowId":    "missing",
		"milestoneId": "missing",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "complete milestone:")
}

func wave54EscrowEngine(settler *wave54Settler) *escrow.Engine {
	cfg := escrow.DefaultEngineConfig()
	cfg.AutoRelease = false
	return escrow.NewEngine(escrow.NewMemoryStore(), settler, cfg)
}

func wave54ActiveEscrow(t *testing.T, engine *escrow.Engine) *escrow.EscrowEntry {
	t.Helper()

	ctx := context.Background()
	entry, err := engine.Create(ctx, escrow.CreateRequest{
		BuyerDID:  "did:leader",
		SellerDID: "did:worker1",
		Amount:    big.NewInt(2500000),
		Reason:    "wave54",
		TaskID:    "task-wave54",
		Milestones: []escrow.MilestoneRequest{
			{Description: "delivery", Amount: big.NewInt(2500000)},
		},
	})
	require.NoError(t, err)
	entry, err = engine.Fund(ctx, entry.ID)
	require.NoError(t, err)
	entry, err = engine.Activate(ctx, entry.ID)
	require.NoError(t, err)
	return entry
}

func requiredParameters(t *testing.T, tool *agent.Tool) []string {
	t.Helper()

	required, ok := tool.Parameters["required"].([]string)
	require.True(t, ok, "required parameters should be []string")
	return required
}

func requireEscrowToolMap(t *testing.T, got interface{}) map[string]interface{} {
	t.Helper()

	result, ok := got.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	return result
}
