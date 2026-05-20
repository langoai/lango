package app

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnChainEscrowTools_RunLifecycleAndQueryViews(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newOnChainEscrowToolsRunLifecycleAndQueryViewsEscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-onChainEscrowToolsRunLifecycleAndQueryViews0", "did:lango:seller-onChainEscrowToolsRunLifecycleAndQueryViews0")

	result, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	assert.Equal(t, "funded", onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)["status"])

	result, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	assert.Equal(t, "active", onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)["status"])

	result, err = rig.tool("escrow_status").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	status := onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)
	assert.Equal(t, "did:lango:buyer-onChainEscrowToolsRunLifecycleAndQueryViews0", status["buyerDid"])
	assert.Equal(t, "did:lango:seller-onChainEscrowToolsRunLifecycleAndQueryViews0", status["sellerDid"])
	assert.Equal(t, "10.00", status["amount"])
	assert.Equal(t, "on-chain delivery proof", status["reason"])
	assert.Equal(t, "active", status["status"])
	milestones, ok := status["milestones"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, milestones, 2)
	assert.Equal(t, "Phase 1", milestones[0]["description"])
	assert.Equal(t, "5.00", milestones[0]["amount"])

	result, err = rig.tool("escrow_submit_work").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"workHash": "proof-hash",
	})
	require.NoError(t, err)
	submit := onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)
	assert.Equal(t, escrowID, submit["escrowId"])
	assert.Equal(t, "active", submit["status"])
	assert.Equal(t, "proof-hash", submit["workHash"])

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"filter": "active"})
	require.NoError(t, err)
	activeList := onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)
	assert.Equal(t, 1, activeList["count"])

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"peerDid": "did:lango:seller-onChainEscrowToolsRunLifecycleAndQueryViews0"})
	require.NoError(t, err)
	peerList := onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)
	assert.Equal(t, 1, peerList["count"])
	items, ok := peerList["escrows"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, items, 1)
	assert.Equal(t, escrowID, items[0]["escrowId"])
}

func TestOnChainEscrowTools_ReleaseAfterMilestonesCompleted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newOnChainEscrowToolsRunLifecycleAndQueryViewsEscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-release", "did:lango:seller-release")
	_, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	_, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)

	entry, err := rig.engine.Get(escrowID)
	require.NoError(t, err)
	for _, milestone := range entry.Milestones {
		entry, err = rig.engine.CompleteMilestone(ctx, escrowID, milestone.ID, "accepted")
		require.NoError(t, err)
	}
	require.Equal(t, escrow.StatusCompleted, entry.Status)

	result, err := rig.tool("escrow_release").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	assert.Equal(t, "released", onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)["status"])
}

func TestOnChainEscrowTools_DisputeRefundAndResolve(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newOnChainEscrowToolsRunLifecycleAndQueryViewsEscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-dispute", "did:lango:seller-dispute")
	_, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	_, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)

	result, err := rig.tool("escrow_dispute").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"note":     "delivery mismatch",
	})
	require.NoError(t, err)
	assert.Equal(t, "disputed", onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)["status"])

	result, err = rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
		"escrowId":      escrowID,
		"favor":         "seller",
		"sellerPercent": float64(100),
	})
	require.NoError(t, err)
	resolve := onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)
	assert.Equal(t, escrowID, resolve["escrowId"])
	assert.Equal(t, "seller", resolve["favor"])
	assert.Equal(t, "released", resolve["status"])
	assert.Equal(t, "10.00", resolve["sellerAmount"])
	assert.Equal(t, "0.00", resolve["buyerAmount"])

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"filter": "disputed"})
	require.NoError(t, err)
	assert.Equal(t, 0, onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)["count"])

	result, err = rig.tool("escrow_refund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "invalid status transition")
}

func TestEscrowResolveTool_RejectsSellerPercentOutsideRange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newOnChainEscrowToolsRunLifecycleAndQueryViewsEscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-percent", "did:lango:seller-percent")

	got, err := rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
		"escrowId":      escrowID,
		"favor":         "seller",
		"sellerPercent": float64(101),
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "sellerPercent must be between 0 and 100")
}

type onChainEscrowToolsRunLifecycleAndQueryViewsEscrowToolRig struct {
	engine *escrow.Engine
	tools  map[string]*agent.Tool
}

func newOnChainEscrowToolsRunLifecycleAndQueryViewsEscrowToolRig() onChainEscrowToolsRunLifecycleAndQueryViewsEscrowToolRig {
	store := escrow.NewMemoryStore()
	settler := &testSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := make(map[string]*agent.Tool)
	for _, tool := range buildOnChainEscrowTools(engine, settler) {
		tools[tool.Name] = tool
	}
	return onChainEscrowToolsRunLifecycleAndQueryViewsEscrowToolRig{engine: engine, tools: tools}
}

func (r onChainEscrowToolsRunLifecycleAndQueryViewsEscrowToolRig) tool(name string) *agent.Tool {
	return r.tools[name]
}

func (r onChainEscrowToolsRunLifecycleAndQueryViewsEscrowToolRig) createEscrow(t *testing.T, ctx context.Context, buyerDID, sellerDID string) string {
	t.Helper()

	result, err := r.tool("escrow_create").Handler(ctx, map[string]interface{}{
		"buyerDid":  buyerDID,
		"sellerDid": sellerDID,
		"amount":    "10.00",
		"reason":    "on-chain delivery proof",
		"milestones": []interface{}{
			map[string]interface{}{"description": "Phase 1", "amount": "5.00"},
			map[string]interface{}{"description": "Phase 2", "amount": "5.00"},
		},
	})
	require.NoError(t, err)
	payload := onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t, result)
	escrowID, ok := payload["escrowId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, escrowID)
	return escrowID
}

func onChainEscrowToolsRunLifecycleAndQueryViewsEscrowPayload(t *testing.T, result interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	return payload
}
