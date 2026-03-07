package app

import (
	"context"
	"math/big"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSettler satisfies escrow.SettlementExecutor with no-op operations for tests.
type testSettler struct{}

func (s *testSettler) Lock(_ context.Context, _ string, _ *big.Int) error   { return nil }
func (s *testSettler) Release(_ context.Context, _ string, _ *big.Int) error { return nil }
func (s *testSettler) Refund(_ context.Context, _ string, _ *big.Int) error  { return nil }

var _ escrow.SettlementExecutor = (*testSettler)(nil)

func TestBuildOnChainEscrowTools(t *testing.T) {
	t.Parallel()

	store := escrow.NewMemoryStore()
	settler := &testSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := buildOnChainEscrowTools(engine, settler)

	assert.Len(t, tools, 10)

	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}

	wantNames := []string{
		"escrow_create",
		"escrow_fund",
		"escrow_activate",
		"escrow_submit_work",
		"escrow_release",
		"escrow_refund",
		"escrow_dispute",
		"escrow_resolve",
		"escrow_status",
		"escrow_list",
	}
	for _, name := range wantNames {
		assert.Contains(t, names, name)
	}
}

func TestBuildOnChainEscrowTools_SafetyLevels(t *testing.T) {
	t.Parallel()

	store := escrow.NewMemoryStore()
	settler := &testSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := buildOnChainEscrowTools(engine, settler)

	toolMap := make(map[string]*agent.Tool, len(tools))
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}

	tests := []struct {
		give     string
		wantSafe bool
	}{
		{give: "escrow_create", wantSafe: false},
		{give: "escrow_fund", wantSafe: false},
		{give: "escrow_activate", wantSafe: false},
		{give: "escrow_submit_work", wantSafe: false},
		{give: "escrow_release", wantSafe: false},
		{give: "escrow_refund", wantSafe: false},
		{give: "escrow_dispute", wantSafe: false},
		{give: "escrow_resolve", wantSafe: false},
		{give: "escrow_status", wantSafe: true},
		{give: "escrow_list", wantSafe: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			tool, ok := toolMap[tt.give]
			require.True(t, ok, "tool %q not found", tt.give)
			isSafe := tool.SafetyLevel == agent.SafetyLevelSafe
			assert.Equal(t, tt.wantSafe, isSafe)
		})
	}
}

func TestEscrowCreateTool_Handler(t *testing.T) {
	t.Parallel()

	store := escrow.NewMemoryStore()
	settler := &testSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := buildOnChainEscrowTools(engine, settler)

	var createTool *agent.Tool
	for _, tool := range tools {
		if tool.Name == "escrow_create" {
			createTool = tool
			break
		}
	}
	require.NotNil(t, createTool)

	result, err := createTool.Handler(context.Background(), map[string]interface{}{
		"buyerDid":  "did:lango:buyer123",
		"sellerDid": "did:lango:seller456",
		"amount":    "10.00",
		"reason":    "Test escrow",
		"milestones": []interface{}{
			map[string]interface{}{"description": "Phase 1", "amount": "5.00"},
			map[string]interface{}{"description": "Phase 2", "amount": "5.00"},
		},
	})
	require.NoError(t, err)

	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, m["escrowId"])
	assert.Equal(t, "pending", m["status"])
	assert.Equal(t, "10.00", m["amount"])
}

func TestEscrowListTool_Handler(t *testing.T) {
	t.Parallel()

	store := escrow.NewMemoryStore()
	settler := &testSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := buildOnChainEscrowTools(engine, settler)

	var listTool *agent.Tool
	for _, tool := range tools {
		if tool.Name == "escrow_list" {
			listTool = tool
			break
		}
	}
	require.NotNil(t, listTool)

	// Empty list.
	result, err := listTool.Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, m["count"])
}
