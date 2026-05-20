package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/contract"
	"github.com/langoai/lango/internal/economy/escrow"
	"github.com/langoai/lango/internal/economy/escrow/hub"
)

func TestEscrowHubToolsAttachOnChainFieldsWithoutLiveChain(t *testing.T) {
	ctx := context.Background()
	caller := newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller()
	hubAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	settler := hub.NewHubSettler(
		caller,
		hubAddr,
		common.HexToAddress("0x2222222222222222222222222222222222222222"),
		31337,
	)
	rig := newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig(settler)
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-hub", "did:lango:seller-hub")
	settler.SetDealMapping(escrowID, big.NewInt(27))

	fundResult, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	fund := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, fundResult)
	assert.Equal(t, "funded", fund["status"])
	assert.Equal(t, "27", fund["dealId"])
	assert.NotEmpty(t, fund["onChainTxHash"])
	assert.True(t, caller.sawWrite("deposit"))
	deposit := caller.requireWriteWithBigIntArg(t, "deposit", 0, 27)
	assert.Equal(t, int64(31337), deposit.ChainID)
	assert.Equal(t, hubAddr, deposit.Address)
	require.Len(t, deposit.Args, 1)
	assertEscrowHubToolsAttachOnChainFieldsWithoutLiveChainBigIntArg(t, deposit.Args, 0, 27)

	_, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)

	submitResult, err := rig.tool("escrow_submit_work").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"workHash": "buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7-proof",
	})
	require.NoError(t, err)
	submit := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, submitResult)
	assert.Equal(t, "buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7-proof", submit["workHash"])
	assert.Equal(t, "27", submit["dealId"])
	assert.NotEmpty(t, submit["onChainTxHash"])
	assert.True(t, caller.sawWrite("submitWork"))
	submitReq := caller.requireWriteWithBigIntArg(t, "submitWork", 0, 27)
	assert.Equal(t, int64(31337), submitReq.ChainID)
	assert.Equal(t, hubAddr, submitReq.Address)
	require.Len(t, submitReq.Args, 2)
	assertEscrowHubToolsAttachOnChainFieldsWithoutLiveChainBigIntArg(t, submitReq.Args, 0, 27)
	assert.Equal(t, sha256.Sum256([]byte("buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7-proof")), submitReq.Args[1])

	statusResult, err := rig.tool("escrow_status").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	status := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, statusResult)
	assert.Equal(t, "27", status["dealId"])
	assert.Equal(t, "deposited", status["onChainStatus"])
	assert.Equal(t, "12500000", status["onChainAmount"])
	assert.True(t, caller.sawRead("getDeal"))

	disputeResult, err := rig.tool("escrow_dispute").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"note":     "buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7 dispute",
	})
	require.NoError(t, err)
	dispute := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, disputeResult)
	assert.Equal(t, "disputed", dispute["status"])
	assert.Equal(t, "27", dispute["dealId"])
	assert.NotEmpty(t, dispute["onChainTxHash"])
	assert.True(t, caller.sawWrite("dispute"))
	disputeReq := caller.requireWriteWithBigIntArg(t, "dispute", 0, 27)
	assert.Equal(t, int64(31337), disputeReq.ChainID)
	assert.Equal(t, hubAddr, disputeReq.Address)
	require.Len(t, disputeReq.Args, 1)
	assertEscrowHubToolsAttachOnChainFieldsWithoutLiveChainBigIntArg(t, disputeReq.Args, 0, 27)

	resolveResult, err := rig.tool("escrow_resolve").Handler(ctx, map[string]interface{}{
		"escrowId":      escrowID,
		"favor":         "seller",
		"sellerPercent": float64(100),
	})
	require.NoError(t, err)
	resolve := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, resolveResult)
	assert.Equal(t, "released", resolve["status"])
	assert.Equal(t, "12.50", resolve["sellerAmount"])
	assert.Equal(t, "0.00", resolve["buyerAmount"])
	assert.Equal(t, "27", resolve["dealId"])
	assert.NotEmpty(t, resolve["onChainTxHash"])
	assert.True(t, caller.sawWrite("resolveDispute"))
	resolveReq := caller.requireWriteWithBigIntArg(t, "resolveDispute", 0, 27)
	assert.Equal(t, int64(31337), resolveReq.ChainID)
	assert.Equal(t, hubAddr, resolveReq.Address)
	require.Len(t, resolveReq.Args, 4)
	assertEscrowHubToolsAttachOnChainFieldsWithoutLiveChainBigIntArg(t, resolveReq.Args, 0, 27)
	assert.Equal(t, true, resolveReq.Args[1])
	assertEscrowHubToolsAttachOnChainFieldsWithoutLiveChainBigIntArg(t, resolveReq.Args, 2, 12_500_000)
	assertEscrowHubToolsAttachOnChainFieldsWithoutLiveChainBigIntArg(t, resolveReq.Args, 3, 0)
}

func TestEscrowHubReleaseAndRefundUseLocalMappings(t *testing.T) {
	ctx := context.Background()

	t.Run("release", func(t *testing.T) {
		caller := newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller()
		hubAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
		settler := hub.NewHubSettler(
			caller,
			hubAddr,
			common.HexToAddress("0x2222222222222222222222222222222222222222"),
			31337,
		)
		rig := newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig(settler)
		escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-release-hub", "did:lango:seller-release-hub")
		settler.SetDealMapping(escrowID, big.NewInt(31))
		settler.SetDealMappingByDID("did:lango:seller-release-hub", big.NewInt(31))
		_, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
		require.NoError(t, err)
		_, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
		require.NoError(t, err)
		rig.completeAllMilestones(t, ctx, escrowID)

		got, err := rig.tool("escrow_release").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
		require.NoError(t, err)
		payload := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, got)
		assert.Equal(t, "released", payload["status"])
		assert.Equal(t, "31", payload["dealId"])
		assert.NotEmpty(t, payload["onChainTxHash"])
		assert.True(t, caller.sawWrite("release"))
		releaseReq := caller.requireWriteWithBigIntArg(t, "release", 0, 31)
		assert.Equal(t, int64(31337), releaseReq.ChainID)
		assert.Equal(t, hubAddr, releaseReq.Address)
		require.Len(t, releaseReq.Args, 1)
		assertEscrowHubToolsAttachOnChainFieldsWithoutLiveChainBigIntArg(t, releaseReq.Args, 0, 31)
	})

	t.Run("refund", func(t *testing.T) {
		caller := newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller()
		hubAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
		settler := hub.NewHubSettler(
			caller,
			hubAddr,
			common.HexToAddress("0x2222222222222222222222222222222222222222"),
			31337,
		)
		rig := newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig(settler)
		escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-refund-hub", "did:lango:seller-refund-hub")
		settler.SetDealMapping(escrowID, big.NewInt(32))
		settler.SetDealMappingByDID("did:lango:buyer-refund-hub", big.NewInt(32))
		_, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
		require.NoError(t, err)
		_, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
		require.NoError(t, err)
		_, err = rig.tool("escrow_dispute").Handler(ctx, map[string]interface{}{
			"escrowId": escrowID,
			"note":     "buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7 refund dispute",
		})
		require.NoError(t, err)

		got, err := rig.tool("escrow_refund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
		require.NoError(t, err)
		payload := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, got)
		assert.Equal(t, "refunded", payload["status"])
		assert.Equal(t, "32", payload["dealId"])
		assert.NotEmpty(t, payload["onChainTxHash"])
		assert.True(t, caller.sawWrite("refund"))
		refundReq := caller.requireWriteWithBigIntArg(t, "refund", 0, 32)
		assert.Equal(t, int64(31337), refundReq.ChainID)
		assert.Equal(t, hubAddr, refundReq.Address)
		require.Len(t, refundReq.Args, 1)
		assertEscrowHubToolsAttachOnChainFieldsWithoutLiveChainBigIntArg(t, refundReq.Args, 0, 32)
	})
}

func TestEscrowVaultToolsAttachOnChainFieldsWithoutLiveChain(t *testing.T) {
	ctx := context.Background()
	caller := newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller()
	vaultAddr := common.HexToAddress("0x3333333333333333333333333333333333333333")
	settler := hub.NewVaultSettler(
		caller,
		common.HexToAddress("0x4444444444444444444444444444444444444444"),
		common.HexToAddress("0x5555555555555555555555555555555555555555"),
		common.HexToAddress("0x6666666666666666666666666666666666666666"),
		common.HexToAddress("0x7777777777777777777777777777777777777777"),
		31337,
	)
	rig := newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig(settler)
	escrowID := rig.createEscrow(t, ctx, "did:lango:buyer-vault", "did:lango:seller-vault")
	settler.SetVaultMapping(escrowID, vaultAddr)

	fundResult, err := rig.tool("escrow_fund").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	fund := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, fundResult)
	assert.Equal(t, "funded", fund["status"])
	assert.Equal(t, vaultAddr.Hex(), fund["vaultAddress"])
	assert.NotEmpty(t, fund["onChainTxHash"])
	assert.True(t, caller.sawWrite("deposit"))
	deposit := caller.requireWrite(t, "deposit")
	assert.Equal(t, int64(31337), deposit.ChainID)
	assert.Equal(t, vaultAddr, deposit.Address)
	assert.Empty(t, deposit.Args)

	_, err = rig.tool("escrow_activate").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	submitResult, err := rig.tool("escrow_submit_work").Handler(ctx, map[string]interface{}{
		"escrowId": escrowID,
		"workHash": "buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7-vault-proof",
	})
	require.NoError(t, err)
	submit := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, submitResult)
	assert.Equal(t, vaultAddr.Hex(), submit["vaultAddress"])
	assert.NotEmpty(t, submit["onChainTxHash"])
	assert.True(t, caller.sawWrite("submitWork"))
	submitReq := caller.requireWrite(t, "submitWork")
	assert.Equal(t, int64(31337), submitReq.ChainID)
	assert.Equal(t, vaultAddr, submitReq.Address)
	require.Len(t, submitReq.Args, 1)
	assert.Equal(t, sha256.Sum256([]byte("buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7-vault-proof")), submitReq.Args[0])

	statusResult, err := rig.tool("escrow_status").Handler(ctx, map[string]interface{}{"escrowId": escrowID})
	require.NoError(t, err)
	status := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, statusResult)
	assert.Equal(t, vaultAddr.Hex(), status["vaultAddress"])
	assert.Equal(t, "work_submitted", status["onChainStatus"])
	assert.Equal(t, "12500000", status["onChainAmount"])
	assert.True(t, caller.sawRead("status"))
	assert.True(t, caller.sawRead("amount"))
	statusReq := caller.requireRead(t, "status")
	assert.Equal(t, vaultAddr, statusReq.Address)
	amountReq := caller.requireRead(t, "amount")
	assert.Equal(t, vaultAddr, amountReq.Address)
}

func TestEscrowValidationBranchesRejectMissingRequiredInputs(t *testing.T) {
	t.Parallel()

	rig := newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig(escrow.NoopSettler{})
	tests := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name: "create missing milestones",
			tool: "escrow_create",
			params: map[string]interface{}{
				"buyerDid":  "did:lango:buyer",
				"sellerDid": "did:lango:seller",
				"amount":    "1.00",
			},
			wantErr: "missing milestones parameter",
		},
		{
			name:    "submit work missing work hash",
			tool:    "escrow_submit_work",
			params:  map[string]interface{}{"escrowId": "escrow-1"},
			wantErr: "missing workHash parameter",
		},
		{
			name:    "dispute missing note",
			tool:    "escrow_dispute",
			params:  map[string]interface{}{"escrowId": "escrow-1"},
			wantErr: "missing note parameter",
		},
		{
			name:    "resolve missing escrow id",
			tool:    "escrow_resolve",
			params:  map[string]interface{}{"favor": "seller", "sellerPercent": float64(100)},
			wantErr: "missing escrowId parameter",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := rig.tool(tt.tool).Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

type escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig struct {
	engine *escrow.Engine
	tools  map[string]*agent.Tool
}

func newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig(settler escrow.SettlementExecutor) escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig {
	engine := escrow.NewEngine(escrow.NewMemoryStore(), settler, escrow.DefaultEngineConfig())
	tools := make(map[string]*agent.Tool)
	for _, tool := range buildOnChainEscrowTools(engine, settler) {
		tools[tool.Name] = tool
	}
	return escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig{engine: engine, tools: tools}
}

func (r escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig) tool(name string) *agent.Tool {
	return r.tools[name]
}

func (r escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig) createEscrow(t *testing.T, ctx context.Context, buyerDID, sellerDID string) string {
	t.Helper()

	result, err := r.tool("escrow_create").Handler(ctx, map[string]interface{}{
		"buyerDid":  buyerDID,
		"sellerDid": sellerDID,
		"amount":    "12.50",
		"reason":    "hub escrow delivery",
		"milestones": []interface{}{
			map[string]interface{}{"description": "Draft", "amount": "6.25"},
			map[string]interface{}{"description": "Final", "amount": "6.25"},
		},
	})
	require.NoError(t, err)
	payload := escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t, result)
	escrowID, ok := payload["escrowId"].(string)
	require.True(t, ok)
	require.NotEmpty(t, escrowID)
	return escrowID
}

func (r escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowToolRig) completeAllMilestones(t *testing.T, ctx context.Context, escrowID string) {
	t.Helper()

	entry, err := r.engine.Get(escrowID)
	require.NoError(t, err)
	for _, milestone := range entry.Milestones {
		_, err = r.engine.CompleteMilestone(ctx, escrowID, milestone.ID, "buildMetaToolsWithEscrowEngineRegistersEngineBackedHelpers7 accepted")
		require.NoError(t, err)
	}
}

func escrowHubToolsAttachOnChainFieldsWithoutLiveChainEscrowPayload(t *testing.T, result interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	return payload
}

type escrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller struct {
	mu     sync.Mutex
	writes []contract.ContractCallRequest
	reads  []contract.ContractCallRequest
	nextTx int
}

func newEscrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller() *escrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller {
	return &escrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller{}
}

func (c *escrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller) Read(_ context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error) {
	c.mu.Lock()
	c.reads = append(c.reads, req)
	c.mu.Unlock()

	switch req.Method {
	case "getDeal":
		return &contract.ContractCallResult{Data: []interface{}{struct {
			Buyer    common.Address
			Seller   common.Address
			Token    common.Address
			Amount   *big.Int
			Deadline *big.Int
			Status   uint8
			WorkHash [32]byte
		}{
			Buyer:    common.HexToAddress("0x8888888888888888888888888888888888888888"),
			Seller:   common.HexToAddress("0x9999999999999999999999999999999999999999"),
			Token:    common.HexToAddress("0x2222222222222222222222222222222222222222"),
			Amount:   big.NewInt(12_500_000),
			Deadline: big.NewInt(1_900_000_000),
			Status:   uint8(hub.DealStatusDeposited),
		}}}, nil
	case "status":
		return &contract.ContractCallResult{Data: []interface{}{uint8(hub.DealStatusWorkSubmitted)}}, nil
	case "amount":
		return &contract.ContractCallResult{Data: []interface{}{big.NewInt(12_500_000)}}, nil
	default:
		return nil, fmt.Errorf("unexpected read method %q", req.Method)
	}
}

func (c *escrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller) Write(_ context.Context, req contract.ContractCallRequest) (*contract.ContractCallResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nextTx++
	c.writes = append(c.writes, req)
	result := &contract.ContractCallResult{TxHash: fmt.Sprintf("0xescrow%s%02d", req.Method, c.nextTx)}
	if req.Method == "createDeal" {
		result.Data = []interface{}{big.NewInt(99)}
	}
	return result, nil
}

func (c *escrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller) sawWrite(method string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, got := range c.writes {
		if got.Method == method {
			return true
		}
	}
	return false
}

func (c *escrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller) sawRead(method string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, got := range c.reads {
		if got.Method == method {
			return true
		}
	}
	return false
}

func (c *escrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller) requireWrite(t *testing.T, method string) contract.ContractCallRequest {
	t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, got := range c.writes {
		if got.Method == method {
			return got
		}
	}
	require.FailNowf(t, "missing contract write", "method %q was not called", method)
	return contract.ContractCallRequest{}
}

func (c *escrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller) requireRead(t *testing.T, method string) contract.ContractCallRequest {
	t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, got := range c.reads {
		if got.Method == method {
			return got
		}
	}
	require.FailNowf(t, "missing contract read", "method %q was not called", method)
	return contract.ContractCallRequest{}
}

func (c *escrowHubToolsAttachOnChainFieldsWithoutLiveChainContractCaller) requireWriteWithBigIntArg(t *testing.T, method string, index int, want int64) contract.ContractCallRequest {
	t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, got := range c.writes {
		if got.Method != method || len(got.Args) <= index {
			continue
		}
		arg, ok := got.Args[index].(*big.Int)
		if ok && arg.Int64() == want {
			return got
		}
	}
	require.FailNowf(t, "missing contract write", "method %q with arg[%d]=%d was not called", method, index, want)
	return contract.ContractCallRequest{}
}

func assertEscrowHubToolsAttachOnChainFieldsWithoutLiveChainBigIntArg(t *testing.T, args []interface{}, index int, want int64) {
	t.Helper()

	require.Greater(t, len(args), index)
	got, ok := args[index].(*big.Int)
	require.Truef(t, ok, "arg %d has type %T, want *big.Int", index, args[index])
	assert.Equal(t, want, got.Int64())
}
