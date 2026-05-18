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
	rig := newWave10EscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-wave10", "did:lango:seller-wave10")

	result, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	assert.Equal(t, "funded", wave10EscrowPayload(t, result)["status"])

	result, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	assert.Equal(t, "active", wave10EscrowPayload(t, result)["status"])

	result, err = rig.tool("escrow_status").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	status := wave10EscrowPayload(t, result)
	assert.Equal(t, "did:lango:buyer-wave10", status["buyerDid"])
	assert.Equal(t, "did:lango:seller-wave10", status["sellerDid"])
	assert.Equal(t, "10.00", status["amount"])
	assert.Equal(t, "wave 10 delivery", status["reason"])
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
	submit := wave10EscrowPayload(t, result)
	assert.Equal(t, escrowID, submit["escrowId"])
	assert.Equal(t, "active", submit["status"])
	assert.Equal(t, "proof-hash", submit["workHash"])

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"filter": "active"})
	require.NoError(t, err)
	activeList := wave10EscrowPayload(t, result)
	assert.Equal(t, 1, activeList["count"])

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"peerDid": "did:lango:seller-wave10"})
	require.NoError(t, err)
	peerList := wave10EscrowPayload(t, result)
	assert.Equal(t, 1, peerList["count"])
	items, ok := peerList["escrows"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, items, 1)
	assert.Equal(t, escrowID, items[0]["escrowId"])
}

func TestOnChainEscrowTools_ReleaseAfterMilestonesCompleted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newWave10EscrowToolRig()
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
	assert.Equal(t, "released", wave10EscrowPayload(t, result)["status"])
}

func TestOnChainEscrowTools_DisputeRefundAndResolve(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newWave10EscrowToolRig()
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
	assert.Equal(t, "disputed", wave10EscrowPayload(t, result)["status"])

	result, err = rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
		"escrowId":      escrowID,
		"favor":         "seller",
		"sellerPercent": float64(40),
	})
	require.NoError(t, err)
	resolve := wave10EscrowPayload(t, result)
	assert.Equal(t, escrowID, resolve["escrowId"])
	assert.Equal(t, "seller", resolve["favor"])
	assert.Equal(t, "4.00", resolve["sellerAmount"])
	assert.Equal(t, "6.00", resolve["buyerAmount"])

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"filter": "disputed"})
	require.NoError(t, err)
	assert.Equal(t, 1, wave10EscrowPayload(t, result)["count"])

	result, err = rig.tool("escrow_refund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	assert.Equal(t, "refunded", wave10EscrowPayload(t, result)["status"])
}

func TestEscrowResolveTool_RejectsSellerPercentOutsideRange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newWave10EscrowToolRig()
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

type wave10EscrowToolRig struct {
	engine *escrow.Engine
	tools  map[string]*agent.Tool
}

func newWave10EscrowToolRig() wave10EscrowToolRig {
	store := escrow.NewMemoryStore()
	settler := &testSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := make(map[string]*agent.Tool)
	for _, tool := range buildOnChainEscrowTools(engine, settler) {
		tools[tool.Name] = tool
	}
	return wave10EscrowToolRig{engine: engine, tools: tools}
}

func (r wave10EscrowToolRig) tool(name string) *agent.Tool {
	return r.tools[name]
}

func (r wave10EscrowToolRig) createEscrow(t *testing.T, ctx context.Context, buyerDID, sellerDID string) string {
	t.Helper()

	result, err := r.tool("escrow_create").Handler(ctx, map[string]interface{}{
		"buyerDid":  buyerDID,
		"sellerDid": sellerDID,
		"amount":    "10.00",
		"reason":    "wave 10 delivery",
		"milestones": []interface{}{
			map[string]interface{}{"description": "Phase 1", "amount": "5.00"},
			map[string]interface{}{"description": "Phase 2", "amount": "5.00"},
		},
	})
	require.NoError(t, err)
	payload := wave10EscrowPayload(t, result)
	escrowID, ok := payload["escrowId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, escrowID)
	return escrowID
}

func wave10EscrowPayload(t *testing.T, result interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	return payload
}
