package economy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/economy/escrow"
)

func TestBuildTools_AllNil(t *testing.T) {
	tools := BuildTools(nil, nil, nil, nil, nil)
	assert.Empty(t, tools, "all-nil engines should produce zero tools")
}

func TestBuildTools_ToolNames(t *testing.T) {
	// We cannot easily construct real engines without full store setup,
	// but we verify nil-guard branches produce no panics and the expected
	// tool count/names when engines are non-nil is tested via the
	// integration path in app/ package tests. Here we verify nil handling.

	tests := []struct {
		give     string
		wantZero bool
	}{
		{give: "all nil", wantZero: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			tools := BuildTools(nil, nil, nil, nil, nil)
			if tt.wantZero {
				assert.Empty(t, tools)
			}
		})
	}
}

func TestBuildTools_NoAppImport(t *testing.T) {
	// Compile-time guarantee: this file is in package economy, not app.
	// If economy/tools.go imported app, this test file would not compile
	// in the economy package. This is a structural smoke test.
	_ = BuildTools
}

func findEconomyTool(tools []*agent.Tool, name string) *agent.Tool {
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	return nil
}

func TestEscrowTools_RequireCanonicalInputs(t *testing.T) {
	ee := escrow.NewEngine(escrow.NewMemoryStore(), escrow.NoopSettler{}, escrow.DefaultEngineConfig())
	tools := buildEscrowTools(ee)

	createTool := findEconomyTool(tools, "economy_escrow_create")
	require.NotNil(t, createTool)
	cases := []struct {
		name      string
		params    map[string]interface{}
		wantError string
	}{
		{
			name: "missing buyerDid",
			params: map[string]interface{}{
				"sellerDid":  "did:lango:seller",
				"amount":     "1.00",
				"milestones": []interface{}{map[string]interface{}{"description": "draft", "amount": "1.00"}},
			},
			wantError: "missing buyerDid parameter",
		},
		{
			name: "missing milestones",
			params: map[string]interface{}{
				"buyerDid":  "did:lango:buyer",
				"sellerDid": "did:lango:seller",
				"amount":    "1.00",
			},
			wantError: "missing milestones parameter",
		},
	}
	for _, tt := range cases {
		t.Run("create/"+tt.name, func(t *testing.T) {
			got, err := createTool.Handler(t.Context(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantError)
		})
	}

	milestoneTool := findEconomyTool(tools, "economy_escrow_milestone")
	require.NotNil(t, milestoneTool)
	got, err := milestoneTool.Handler(t.Context(), map[string]interface{}{"escrowId": "escrow-1"})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "missing milestoneId parameter")

	statusTool := findEconomyTool(tools, "economy_escrow_status")
	require.NotNil(t, statusTool)
	got, err = statusTool.Handler(t.Context(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "missing escrowId parameter")

	releaseTool := findEconomyTool(tools, "economy_escrow_release")
	require.NotNil(t, releaseTool)
	got, err = releaseTool.Handler(t.Context(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "missing escrowId parameter")

	disputeTool := findEconomyTool(tools, "economy_escrow_dispute")
	require.NotNil(t, disputeTool)
	got, err = disputeTool.Handler(t.Context(), map[string]interface{}{"escrowId": "escrow-1"})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "missing note parameter")
}

func TestCoreEconomyTools_RequireCanonicalInputs(t *testing.T) {
	tools := append(
		append(buildBudgetTools(nil), buildRiskTools(nil)...),
		append(buildNegotiationTools(nil), buildPricingTools(nil)...)...,
	)

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{name: "budget allocate requires taskId", tool: "economy_budget_allocate", params: map[string]interface{}{}, wantErr: "missing taskId parameter"},
		{name: "budget status requires taskId", tool: "economy_budget_status", params: map[string]interface{}{}, wantErr: "missing taskId parameter"},
		{name: "budget close requires taskId", tool: "economy_budget_close", params: map[string]interface{}{}, wantErr: "missing taskId parameter"},
		{name: "risk assess requires peerDid", tool: "economy_risk_assess", params: map[string]interface{}{"amount": "1.00"}, wantErr: "missing peerDid parameter"},
		{name: "risk assess requires amount", tool: "economy_risk_assess", params: map[string]interface{}{"peerDid": "did:lango:peer"}, wantErr: "missing amount parameter"},
		{name: "negotiate requires peerDid", tool: "economy_negotiate", params: map[string]interface{}{"toolName": "review", "price": "1.00"}, wantErr: "missing peerDid parameter"},
		{name: "negotiate requires toolName", tool: "economy_negotiate", params: map[string]interface{}{"peerDid": "did:lango:peer", "price": "1.00"}, wantErr: "missing toolName parameter"},
		{name: "negotiate requires price", tool: "economy_negotiate", params: map[string]interface{}{"peerDid": "did:lango:peer", "toolName": "review"}, wantErr: "missing price parameter"},
		{name: "negotiate status requires sessionId", tool: "economy_negotiate_status", params: map[string]interface{}{}, wantErr: "missing sessionId parameter"},
		{name: "price quote requires toolName", tool: "economy_price_quote", params: map[string]interface{}{}, wantErr: "missing toolName parameter"},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tool := findEconomyTool(tools, tt.tool)
			require.NotNil(t, tool)
			got, err := tool.Handler(t.Context(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
