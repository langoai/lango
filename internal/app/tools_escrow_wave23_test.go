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

func TestWave23EscrowTools_NonHubLifecycleOmitsOnChainFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newWave23EscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-wave23", "did:lango:seller-wave23")

	result, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	fund := wave23EscrowPayload(t, result)
	assert.Equal(t, escrowID, fund["escrowId"])
	assert.Equal(t, "funded", fund["status"])
	assertWave23NoOnChainFields(t, fund)

	result, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	activate := wave23EscrowPayload(t, result)
	assert.Equal(t, escrowID, activate["escrowId"])
	assert.Equal(t, "active", activate["status"])

	result, err = rig.tool("escrow_submit_work").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"workHash": "wave23-proof",
	})
	require.NoError(t, err)
	submit := wave23EscrowPayload(t, result)
	assert.Equal(t, escrowID, submit["escrowId"])
	assert.Equal(t, "active", submit["status"])
	assert.Equal(t, "wave23-proof", submit["workHash"])
	assertWave23NoOnChainFields(t, submit)

	result, err = rig.tool("escrow_status").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	status := wave23EscrowPayload(t, result)
	assert.Equal(t, "did:lango:buyer-wave23", status["buyerDid"])
	assert.Equal(t, "did:lango:seller-wave23", status["sellerDid"])
	assert.Equal(t, "12.50", status["amount"])
	assert.Equal(t, "active", status["status"])
	assertWave23NoOnChainFields(t, status)

	rig.completeAllMilestones(t, ctx, escrowID)
	result, err = rig.tool("escrow_release").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	release := wave23EscrowPayload(t, result)
	assert.Equal(t, escrowID, release["escrowId"])
	assert.Equal(t, "released", release["status"])
	assertWave23NoOnChainFields(t, release)

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"filter": "active"})
	require.NoError(t, err)
	assert.Equal(t, 0, wave23EscrowPayload(t, result)["count"])

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"peerDid": "did:lango:seller-wave23"})
	require.NoError(t, err)
	list := wave23EscrowPayload(t, result)
	assert.Equal(t, 1, list["count"])
	items, ok := list["escrows"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, items, 1)
	assert.Equal(t, escrowID, items[0]["escrowId"])
	assert.Equal(t, "released", items[0]["status"])
}

func TestWave23EscrowTools_DisputeResolveAndRefundNonHubBranches(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newWave23EscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-dispute-wave23", "did:lango:seller-dispute-wave23")
	_, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	_, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)

	result, err := rig.tool("escrow_dispute").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"note":     "wave23 dispute",
	})
	require.NoError(t, err)
	dispute := wave23EscrowPayload(t, result)
	assert.Equal(t, escrowID, dispute["escrowId"])
	assert.Equal(t, "disputed", dispute["status"])
	assertWave23NoOnChainFields(t, dispute)

	result, err = rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
		"escrowId":      escrowID,
		"favor":         "seller",
		"sellerPercent": float64(100),
	})
	require.NoError(t, err)
	resolve := wave23EscrowPayload(t, result)
	assert.Equal(t, escrowID, resolve["escrowId"])
	assert.Equal(t, "seller", resolve["favor"])
	assert.Equal(t, "released", resolve["status"])
	assert.Equal(t, "12.50", resolve["sellerAmount"])
	assert.Equal(t, "0.00", resolve["buyerAmount"])
	assertWave23NoOnChainFields(t, resolve)
	require.Len(t, rig.settler.released, 1)
	assert.Empty(t, rig.settler.refunded)
	assert.Equal(t, "12500000", rig.settler.released[0].String())

	result, err = rig.tool("escrow_list").Handler(ctx, map[string]interface{}{"filter": "disputed"})
	require.NoError(t, err)
	assert.Equal(t, 0, wave23EscrowPayload(t, result)["count"])

	result, err = rig.tool("escrow_refund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "invalid status transition")
}

func TestWave23EscrowResolveTool_BuyerFavorRefundsDispute(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newWave23EscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-refund-wave23", "did:lango:seller-refund-wave23")
	_, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	_, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	_, err = rig.tool("escrow_dispute").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"note":     "wave23 buyer dispute",
	})
	require.NoError(t, err)

	result, err := rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
		"escrowId":      escrowID,
		"favor":         "buyer",
		"sellerPercent": float64(0),
	})
	require.NoError(t, err)
	refund := wave23EscrowPayload(t, result)
	assert.Equal(t, escrowID, refund["escrowId"])
	assert.Equal(t, "buyer", refund["favor"])
	assert.Equal(t, "refunded", refund["status"])
	assert.Equal(t, "0.00", refund["sellerAmount"])
	assert.Equal(t, "12.50", refund["buyerAmount"])
	assertWave23NoOnChainFields(t, refund)
	assert.Empty(t, rig.settler.released)
	require.Len(t, rig.settler.refunded, 1)
	assert.Equal(t, "12500000", rig.settler.refunded[0].String())
}

func TestWave23EscrowResolveTool_RejectsPartialFavorSplits(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newWave23EscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-split-wave23", "did:lango:seller-split-wave23")

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

func TestWave23EscrowResolveTool_RejectsNonDisputedBeforeSettlement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newWave23EscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-nondisputed-wave23", "did:lango:seller-nondisputed-wave23")
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

func TestWave23EscrowResolveTool_RejectsUnknownFavor(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newWave23EscrowToolRig()
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-favor-wave23", "did:lango:seller-favor-wave23")

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

func TestWave23EscrowTools_CreateValidationAndEngineErrors(t *testing.T) {
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

			rig := newWave23EscrowToolRig()
			got, err := rig.tool("escrow_create").Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestWave23EscrowTools_StateAndLookupErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rig := newWave23EscrowToolRig()
	pendingEscrowID := rig.createEscrow(t, ctx, "did:lango:buyer-state-wave23", "did:lango:seller-state-wave23")

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

func TestWave23EscrowResolveTool_ValidatesSellerPercent(t *testing.T) {
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

			rig := newWave23EscrowToolRig()
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

type wave23EscrowToolRig struct {
	engine  *escrow.Engine
	settler *wave23RecordingSettler
	tools   map[string]*agent.Tool
}

func newWave23EscrowToolRig() wave23EscrowToolRig {
	store := escrow.NewMemoryStore()
	settler := &wave23RecordingSettler{}
	engine := escrow.NewEngine(store, settler, escrow.DefaultEngineConfig())
	tools := make(map[string]*agent.Tool)
	for _, tool := range buildOnChainEscrowTools(engine, settler) {
		tools[tool.Name] = tool
	}
	return wave23EscrowToolRig{engine: engine, settler: settler, tools: tools}
}

type wave23RecordingSettler struct {
	locked   []*big.Int
	released []*big.Int
	refunded []*big.Int
}

func (s *wave23RecordingSettler) Lock(_ context.Context, _ string, amount *big.Int) error {
	s.locked = append(s.locked, new(big.Int).Set(amount))
	return nil
}

func (s *wave23RecordingSettler) Release(_ context.Context, _ string, amount *big.Int) error {
	s.released = append(s.released, new(big.Int).Set(amount))
	return nil
}

func (s *wave23RecordingSettler) Refund(_ context.Context, _ string, amount *big.Int) error {
	s.refunded = append(s.refunded, new(big.Int).Set(amount))
	return nil
}

func (r wave23EscrowToolRig) tool(name string) *agent.Tool {
	return r.tools[name]
}

func (r wave23EscrowToolRig) createEscrow(t *testing.T, ctx context.Context, buyerDID, sellerDID string) string {
	t.Helper()

	result, err := r.tool("escrow_create").Handler(ctx, map[string]interface{}{
		"buyerDid":  buyerDID,
		"sellerDid": sellerDID,
		"amount":    "12.50",
		"reason":    "wave 23 delivery",
		"milestones": []interface{}{
			map[string]interface{}{"description": "Draft", "amount": "6.25"},
			map[string]interface{}{"description": "Final", "amount": "6.25"},
		},
	})
	require.NoError(t, err)
	payload := wave23EscrowPayload(t, result)
	escrowID, ok := payload["escrowId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, escrowID)
	return escrowID
}

func (r wave23EscrowToolRig) completeAllMilestones(t *testing.T, ctx context.Context, escrowID string) {
	t.Helper()

	entry, err := r.engine.Get(escrowID)
	require.NoError(t, err)
	for _, milestone := range entry.Milestones {
		_, err = r.engine.CompleteMilestone(ctx, escrowID, milestone.ID, "wave23 accepted")
		require.NoError(t, err)
	}
}

func wave23EscrowPayload(t *testing.T, result interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	return payload
}

func assertWave23NoOnChainFields(t *testing.T, payload map[string]interface{}) {
	t.Helper()

	assert.NotContains(t, payload, "onChainTxHash")
	assert.NotContains(t, payload, "dealId")
	assert.NotContains(t, payload, "vaultAddress")
	assert.NotContains(t, payload, "onChainStatus")
	assert.NotContains(t, payload, "onChainAmount")
}
