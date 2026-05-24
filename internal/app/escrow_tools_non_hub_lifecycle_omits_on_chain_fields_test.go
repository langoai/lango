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

func TestEscrowTools_NonHubLifecycleOmitsOnChainFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newEscrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-escrowToolsNonHubLifecycleOmitsOnChainFields3", "did:lango:seller-escrowToolsNonHubLifecycleOmitsOnChainFields3")

	result, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	fund := escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)
	assert.Equal(t, escrowID, fund["escrowId"])
	assert.Equal(t, "funded", fund["status"])
	assertEscrowToolsNonHubLifecycleOmitsOnChainFieldsNoOnChainFields(t, fund)

	result, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	activate := escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)
	assert.Equal(t, escrowID, activate["escrowId"])
	assert.Equal(t, "active", activate["status"])

	result, err = rig.tool("escrow_submit_work").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"workHash": "escrowToolsNonHubLifecycleOmitsOnChainFields3-proof",
	})
	require.NoError(t, err)
	submit := escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)
	assert.Equal(t, escrowID, submit["escrowId"])
	assert.Equal(t, "active", submit["status"])
	assert.Equal(t, "escrowToolsNonHubLifecycleOmitsOnChainFields3-proof", submit["workHash"])
	assertEscrowToolsNonHubLifecycleOmitsOnChainFieldsNoOnChainFields(t, submit)

	result, err = rig.tool("escrow_status").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	status := escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)
	assert.Equal(t, "did:lango:buyer-escrowToolsNonHubLifecycleOmitsOnChainFields3", status["buyerDid"])
	assert.Equal(t, "did:lango:seller-escrowToolsNonHubLifecycleOmitsOnChainFields3", status["sellerDid"])
	assert.Equal(t, "12.50", status["amount"])
	assert.Equal(t, "active", status["status"])
	assertEscrowToolsNonHubLifecycleOmitsOnChainFieldsNoOnChainFields(t, status)

	rig.completeAllMilestones(t, ctx, escrowID)
	result, err = rig.tool("escrow_release").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	release := escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)
	assert.Equal(t, escrowID, release["escrowId"])
	assert.Equal(t, "released", release["status"])
	assertEscrowToolsNonHubLifecycleOmitsOnChainFieldsNoOnChainFields(t, release)

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"filter": "active"})
	require.NoError(t, err)
	assert.Equal(t, 0, escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)["count"])

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"peerDid": "did:lango:seller-escrowToolsNonHubLifecycleOmitsOnChainFields3"})
	require.NoError(t, err)
	list := escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)
	assert.Equal(t, 1, list["count"])
	items, ok := list["escrows"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, items, 1)
	assert.Equal(t, escrowID, items[0]["escrowId"])
	assert.Equal(t, "released", items[0]["status"])
}

func TestEscrowTools_DisputeResolveAndRefundNonHubBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newEscrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-dispute-escrowToolsNonHubLifecycleOmitsOnChainFields3", "did:lango:seller-dispute-escrowToolsNonHubLifecycleOmitsOnChainFields3")
	_, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	_, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)

	result, err := rig.tool("escrow_dispute").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"note":     "escrowToolsNonHubLifecycleOmitsOnChainFields3 dispute",
	})
	require.NoError(t, err)
	dispute := escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)
	assert.Equal(t, escrowID, dispute["escrowId"])
	assert.Equal(t, "disputed", dispute["status"])
	assertEscrowToolsNonHubLifecycleOmitsOnChainFieldsNoOnChainFields(t, dispute)

	result, err = rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
		"escrowId":      escrowID,
		"favor":         "seller",
		"sellerPercent": float64(100),
	})
	require.NoError(t, err)
	resolve := escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)
	assert.Equal(t, escrowID, resolve["escrowId"])
	assert.Equal(t, "seller", resolve["favor"])
	assert.Equal(t, "released", resolve["status"])
	assert.Equal(t, "12.50", resolve["sellerAmount"])
	assert.Equal(t, "0.00", resolve["buyerAmount"])
	assertEscrowToolsNonHubLifecycleOmitsOnChainFieldsNoOnChainFields(t, resolve)
	require.Len(t, rig.settler.released, 1)
	assert.Empty(t, rig.settler.refunded)
	assert.Equal(t, "12500000", rig.settler.released[0].String())

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"filter": "disputed"})
	require.NoError(t, err)
	assert.Equal(t, 0, escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)["count"])

	result, err = rig.tool("escrow_refund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "invalid status transition")
}

func TestEscrowResolveTool_BuyerFavorRefundsDispute(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newEscrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-refund-escrowToolsNonHubLifecycleOmitsOnChainFields3", "did:lango:seller-refund-escrowToolsNonHubLifecycleOmitsOnChainFields3")
	_, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	_, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	_, err = rig.tool("escrow_dispute").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"note":     "escrowToolsNonHubLifecycleOmitsOnChainFields3 buyer dispute",
	})
	require.NoError(t, err)

	result, err := rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
		"escrowId":      escrowID,
		"favor":         "buyer",
		"sellerPercent": float64(0),
	})
	require.NoError(t, err)
	refund := escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)
	assert.Equal(t, escrowID, refund["escrowId"])
	assert.Equal(t, "buyer", refund["favor"])
	assert.Equal(t, "refunded", refund["status"])
	assert.Equal(t, "0.00", refund["sellerAmount"])
	assert.Equal(t, "12.50", refund["buyerAmount"])
	assertEscrowToolsNonHubLifecycleOmitsOnChainFieldsNoOnChainFields(t, refund)
	assert.Empty(t, rig.settler.released)
	require.Len(t, rig.settler.refunded, 1)
	assert.Equal(t, "12500000", rig.settler.refunded[0].String())
}

func TestEscrowResolveTool_RejectsPartialFavorSplits(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newEscrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-split-escrowToolsNonHubLifecycleOmitsOnChainFields3", "did:lango:seller-split-escrowToolsNonHubLifecycleOmitsOnChainFields3")

	tests := []struct {
		name          string
		favor         string
		sellerPercent float64
		wantErr       string
	}{
		{
			name:          "buyer favor with seller share",
			favor:         "buyer",
			sellerPercent: 25,
			wantErr:       "sellerPercent must be 0 when favor is buyer",
		},
		{
			name:          "seller favor with buyer share",
			favor:         "seller",
			sellerPercent: 75,
			wantErr:       "sellerPercent must be 100 when favor is seller",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
				"escrowId":      escrowID,
				"favor":         tt.favor,
				"sellerPercent": tt.sellerPercent,
			})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestEscrowResolveTool_RejectsNonDisputedBeforeSettlement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newEscrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-nondisputed-escrowToolsNonHubLifecycleOmitsOnChainFields3", "did:lango:seller-nondisputed-escrowToolsNonHubLifecycleOmitsOnChainFields3")
	_, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)

	got, err := rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
		"escrowId":      escrowID,
		"favor":         "seller",
		"sellerPercent": float64(100),
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, escrow.ErrInvalidTransition)
	assert.Empty(t, rig.settler.released)
	assert.Empty(t, rig.settler.refunded)
}

func TestEscrowResolveTool_RejectsUnknownFavor(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newEscrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-favor-escrowToolsNonHubLifecycleOmitsOnChainFields3", "did:lango:seller-favor-escrowToolsNonHubLifecycleOmitsOnChainFields3")

	got, err := rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
		"escrowId":      escrowID,
		"favor":         "seler",
		"sellerPercent": float64(100),
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "favor must be either buyer or seller")
	assert.Empty(t, rig.settler.released)
	assert.Empty(t, rig.settler.refunded)
}

func TestEscrowTools_CreateValidationAndEngineErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name: "invalid total amount",
			params: map[string]interface{}{
				"buyerDid":  "did:lango:buyer",
				"sellerDid": "did:lango:seller",
				"amount":    "not-usdc",
				"milestones": []interface{}{
					map[string]interface{}{"description": "Phase 1", "amount": "1.00"},
				},
			},
			wantErr: "parse amount",
		},
		{
			name: "invalid milestone amount",
			params: map[string]interface{}{
				"buyerDid":  "did:lango:buyer",
				"sellerDid": "did:lango:seller",
				"amount":    "1.00",
				"milestones": []interface{}{
					map[string]interface{}{"description": "Phase 1", "amount": "bad"},
				},
			},
			wantErr: "parse milestone amount",
		},
		{
			name: "milestone sum mismatch",
			params: map[string]interface{}{
				"buyerDid":  "did:lango:buyer",
				"sellerDid": "did:lango:seller",
				"amount":    "2.00",
				"milestones": []interface{}{
					map[string]interface{}{"description": "Phase 1", "amount": "1.00"},
				},
			},
			wantErr: "milestone amounts do not match total",
		},
		{
			name: "malformed milestone payload reaches no milestone engine error",
			params: map[string]interface{}{
				"buyerDid":   "did:lango:buyer",
				"sellerDid":  "did:lango:seller",
				"amount":     "1.00",
				"milestones": []interface{}{"not-a-map"},
			},
			wantErr: "escrow has no milestones",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rig := newEscrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig()
			got, err := rig.tool("escrow_create").Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestEscrowTools_StateAndLookupErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newEscrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig()
	pendingEscrowID := rig.createEscrow(t, ctx, "did:lango:buyer-state-escrowToolsNonHubLifecycleOmitsOnChainFields3", "did:lango:seller-state-escrowToolsNonHubLifecycleOmitsOnChainFields3")

	tests := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "fund unknown escrow",
			tool:    "escrow_fund",
			params:  map[string]interface{}{"escrowId": "missing-escrow"},
			wantErr: "escrow not found",
		},
		{
			name:    "activate pending escrow",
			tool:    "escrow_activate",
			params:  map[string]interface{}{"escrowId": pendingEscrowID},
			wantErr: "invalid status transition",
		},
		{
			name:    "submit work unknown escrow",
			tool:    "escrow_submit_work",
			params:  map[string]interface{}{"escrowId": "missing-escrow", "workHash": "proof"},
			wantErr: "escrow not found",
		},
		{
			name:    "release pending escrow",
			tool:    "escrow_release",
			params:  map[string]interface{}{"escrowId": pendingEscrowID},
			wantErr: "invalid status transition",
		},
		{
			name:    "refund pending escrow",
			tool:    "escrow_refund",
			params:  map[string]interface{}{"escrowId": pendingEscrowID},
			wantErr: "invalid status transition",
		},
		{
			name:    "dispute pending escrow",
			tool:    "escrow_dispute",
			params:  map[string]interface{}{"escrowId": pendingEscrowID, "note": "too early"},
			wantErr: "invalid status transition",
		},
		{
			name:    "resolve unknown escrow",
			tool:    "escrow_resolve",
			params:  map[string]interface{}{"escrowId": "missing-escrow", "favor": "buyer", "sellerPercent": float64(0)},
			wantErr: "escrow not found",
		},
		{
			name:    "status unknown escrow",
			tool:    "escrow_status",
			params:  map[string]interface{}{"escrowId": "missing-escrow"},
			wantErr: "escrow not found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := rig.tool(tt.tool).Handler(ctx, tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestEscrowResolveTool_ValidatesSellerPercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		sellerPercent interface{}
		wantErr       string
	}{
		{
			name:          "missing numeric parameter when string is supplied",
			sellerPercent: "50",
			wantErr:       "missing sellerPercent parameter",
		},
		{
			name:          "below lower bound",
			sellerPercent: float64(-1),
			wantErr:       "sellerPercent must be between 0 and 100",
		},
		{
			name:          "above upper bound",
			sellerPercent: float64(100.01),
			wantErr:       "sellerPercent must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rig := newEscrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig()
			got, err := rig.tool("escrow_resolve").Handler(context.Background(), map[string]interface{}{
				"escrowId":      "escrow-1",
				"favor":         "seller",
				"sellerPercent": tt.sellerPercent,
			})
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

type escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig struct {
	engine  *escrow.Engine
	settler *escrowToolsNonHubLifecycleOmitsOnChainFieldsRecordingSettler
	tools   map[string]*agent.Tool
}

func newEscrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig() escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig {
	store := escrow.NewMemoryStore()
	settler := &escrowToolsNonHubLifecycleOmitsOnChainFieldsRecordingSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := make(map[string]*agent.Tool)
	for _, tool := range buildOnChainEscrowTools(engine, settler) {
		tools[tool.Name] = tool
	}
	return escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig{engine: engine, settler: settler, tools: tools}
}

type escrowToolsNonHubLifecycleOmitsOnChainFieldsRecordingSettler struct {
	locked   []*big.Int
	released []*big.Int
	refunded []*big.Int
}

func (s *escrowToolsNonHubLifecycleOmitsOnChainFieldsRecordingSettler) Lock(_ context.Context, _ string, amount *big.Int) error {
	s.locked = append(s.locked, new(big.Int).Set(amount))
	return nil
}

func (s *escrowToolsNonHubLifecycleOmitsOnChainFieldsRecordingSettler) Release(_ context.Context, _ string, amount *big.Int) error {
	s.released = append(s.released, new(big.Int).Set(amount))
	return nil
}

func (s *escrowToolsNonHubLifecycleOmitsOnChainFieldsRecordingSettler) Refund(_ context.Context, _ string, amount *big.Int) error {
	s.refunded = append(s.refunded, new(big.Int).Set(amount))
	return nil
}

func (r escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig) tool(name string) *agent.Tool {
	return r.tools[name]
}

func (r escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig) createEscrow(t *testing.T, ctx context.Context, buyerDID, sellerDID string) string {
	t.Helper()

	result, err := r.tool("escrow_create").Handler(ctx, map[string]interface{}{
		"buyerDid":  buyerDID,
		"sellerDid": sellerDID,
		"amount":    "12.50",
		"reason":    "off-chain delivery proof",
		"milestones": []interface{}{
			map[string]interface{}{"description": "Draft", "amount": "6.25"},
			map[string]interface{}{"description": "Final", "amount": "6.25"},
		},
	})
	require.NoError(t, err)
	payload := escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t, result)
	escrowID, ok := payload["escrowId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, escrowID)
	return escrowID
}

func (r escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowToolRig) completeAllMilestones(t *testing.T, ctx context.Context, escrowID string) {
	t.Helper()

	entry, err := r.engine.Get(escrowID)
	require.NoError(t, err)
	for _, milestone := range entry.Milestones {
		_, err = r.engine.CompleteMilestone(ctx, escrowID, milestone.ID, "escrowToolsNonHubLifecycleOmitsOnChainFields3 accepted")
		require.NoError(t, err)
	}
}

func escrowToolsNonHubLifecycleOmitsOnChainFieldsEscrowPayload(t *testing.T, result interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	return payload
}

func assertEscrowToolsNonHubLifecycleOmitsOnChainFieldsNoOnChainFields(t *testing.T, payload map[string]interface{}) {
	t.Helper()

	assert.NotContains(t, payload, "onChainTxHash")
	assert.NotContains(t, payload, "dealId")
	assert.NotContains(t, payload, "vaultAddress")
	assert.NotContains(t, payload, "onChainStatus")
	assert.NotContains(t, payload, "onChainAmount")
}
