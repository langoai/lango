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

func (s *testSettler) Lock(_ context.Context, _ string, _ *big.Int) error    { return nil }
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

func TestEscrowCreateTool_RequiresMilestonesParameter(t *testing.T) {
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

	got, err := createTool.Handler(context.Background(), map[string]interface{}{
		"buyerDid":  "did:lango:buyer123",
		"sellerDid": "did:lango:seller456",
		"amount":    "10.00",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "missing milestones parameter")
}

func TestEscrowCreateTool_RequiresCanonicalInputs(t *testing.T) {
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

	testCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name: "create requires buyerDid",
			params: map[string]interface{}{
				"sellerDid": "did:lango:seller456",
				"amount":    "10.00",
				"milestones": []interface{}{
					map[string]interface{}{"description": "Phase 1", "amount": "5.00"},
				},
			},
			wantErr: "missing buyerDid parameter",
		},
		{
			name: "create requires sellerDid",
			params: map[string]interface{}{
				"buyerDid": "did:lango:buyer123",
				"amount":   "10.00",
				"milestones": []interface{}{
					map[string]interface{}{"description": "Phase 1", "amount": "5.00"},
				},
			},
			wantErr: "missing sellerDid parameter",
		},
		{
			name: "create requires amount",
			params: map[string]interface{}{
				"buyerDid":  "did:lango:buyer123",
				"sellerDid": "did:lango:seller456",
				"milestones": []interface{}{
					map[string]interface{}{"description": "Phase 1", "amount": "5.00"},
				},
			},
			wantErr: "missing amount parameter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := createTool.Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
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

func TestEscrowResolveTool_RequiresSellerPercentParameter(t *testing.T) {
	t.Parallel()

	store := escrow.NewMemoryStore()
	settler := &testSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := buildOnChainEscrowTools(engine, settler)

	var resolveTool *agent.Tool
	for _, tool := range tools {
		if tool.Name == "escrow_resolve" {
			resolveTool = tool
			break
		}
	}
	require.NotNil(t, resolveTool)

	got, err := resolveTool.Handler(context.Background(), map[string]interface{}{
		"escrowId": "escrow-1",
		"favor":    "seller",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "missing sellerPercent parameter")
}

func TestEscrowResolveTool_RequiresFavorParameter(t *testing.T) {
	t.Parallel()

	store := escrow.NewMemoryStore()
	settler := &testSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := buildOnChainEscrowTools(engine, settler)

	var resolveTool *agent.Tool
	for _, tool := range tools {
		if tool.Name == "escrow_resolve" {
			resolveTool = tool
			break
		}
	}
	require.NotNil(t, resolveTool)

	got, err := resolveTool.Handler(context.Background(), map[string]interface{}{
		"escrowId":      "escrow-1",
		"sellerPercent": float64(50),
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "missing favor parameter")
}

func TestOnChainEscrowTools_RequireCanonicalInputs(t *testing.T) {
	store := escrow.NewMemoryStore()
	settler := &testSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := buildOnChainEscrowTools(engine, settler)

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{name: "fund requires escrowId", tool: "escrow_fund", params: map[string]interface{}{}, wantErr: "missing escrowId parameter"},
		{name: "activate requires escrowId", tool: "escrow_activate", params: map[string]interface{}{}, wantErr: "missing escrowId parameter"},
		{name: "submit work requires escrowId", tool: "escrow_submit_work", params: map[string]interface{}{"workHash": "proof"}, wantErr: "missing escrowId parameter"},
		{name: "submit work requires workHash", tool: "escrow_submit_work", params: map[string]interface{}{"escrowId": "escrow-1"}, wantErr: "missing workHash parameter"},
		{name: "release requires escrowId", tool: "escrow_release", params: map[string]interface{}{}, wantErr: "missing escrowId parameter"},
		{name: "refund requires escrowId", tool: "escrow_refund", params: map[string]interface{}{}, wantErr: "missing escrowId parameter"},
		{name: "dispute requires escrowId", tool: "escrow_dispute", params: map[string]interface{}{"note": "problem"}, wantErr: "missing escrowId parameter"},
		{name: "dispute requires note", tool: "escrow_dispute", params: map[string]interface{}{"escrowId": "escrow-1"}, wantErr: "missing note parameter"},
		{name: "status requires escrowId", tool: "escrow_status", params: map[string]interface{}{}, wantErr: "missing escrowId parameter"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var selected *agent.Tool
			for _, tool := range tools {
				if tool.Name == tc.tool {
					selected = tool
					break
				}
			}
			require.NotNil(t, selected)

			got, err := selected.Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}
