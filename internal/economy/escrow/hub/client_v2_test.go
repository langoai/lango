package hub

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/contract"
)

func TestParseHubV2ABI(t *testing.T) {
	abi, err := ParseHubV2ABI()
	require.NoError(t, err)
	require.NotNil(t, abi)

	// Verify V2-specific methods exist.
	wantMethods := []string{
		"directSettle",
		"createSimpleEscrow",
		"createMilestoneEscrow",
		"createTeamEscrow",
		"completeMilestone",
		"releaseMilestone",
		"registerSettler",
		"deposit",
		"release",
		"refund",
		"dispute",
		"resolveDispute",
		"getDeal",
		"getTeamDeal",
		"nextDealId",
	}
	for _, m := range wantMethods {
		_, ok := abi.Methods[m]
		assert.True(t, ok, "missing method: %s", m)
	}

	// Verify V2-specific events exist.
	wantEvents := []string{
		"EscrowOpened",
		"MilestoneReached",
		"DisputeRaised",
		"SettlementFinalized",
		"Deposited",
		"WorkSubmitted",
		"Released",
		"Refunded",
		"SettlerRegistered",
	}
	for _, e := range wantEvents {
		_, ok := abi.Events[e]
		assert.True(t, ok, "missing event: %s", e)
	}
}

func TestParseVaultV2ABI(t *testing.T) {
	abi, err := ParseVaultV2ABI()
	require.NoError(t, err)
	require.NotNil(t, abi)

	wantMethods := []string{
		"initialize",
		"deposit",
		"submitWork",
		"release",
		"refund",
		"dispute",
		"resolve",
		"setSettler",
	}
	for _, m := range wantMethods {
		_, ok := abi.Methods[m]
		assert.True(t, ok, "missing vault V2 method: %s", m)
	}

	wantEvents := []string{
		"VaultInitialized",
		"Deposited",
		"WorkSubmitted",
		"Released",
		"Refunded",
		"Disputed",
		"VaultResolved",
	}
	for _, e := range wantEvents {
		_, ok := abi.Events[e]
		assert.True(t, ok, "missing vault V2 event: %s", e)
	}
}

func TestOnChainDealType_String(t *testing.T) {
	tests := []struct {
		give OnChainDealType
		want string
	}{
		{DealTypeSimple, "simple"},
		{DealTypeMilestone, "milestone"},
		{DealTypeTeam, "team"},
		{OnChainDealType(99), "unknown"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.give.String())
	}
}

func TestEscrowOnChainConfig_IsV2(t *testing.T) {
	tests := []struct {
		give   string
		config EscrowOnChainConfigForTest
		want   bool
	}{
		{
			give:   "explicit v2",
			config: EscrowOnChainConfigForTest{ContractVersion: "v2"},
			want:   true,
		},
		{
			give:   "explicit v1",
			config: EscrowOnChainConfigForTest{ContractVersion: "v1"},
			want:   false,
		},
		{
			give:   "auto-detect v2 by hub address",
			config: EscrowOnChainConfigForTest{HubV2Address: "0x123"},
			want:   true,
		},
		{
			give:   "auto-detect v2 by beacon factory",
			config: EscrowOnChainConfigForTest{BeaconFactoryAddress: "0x456"},
			want:   true,
		},
		{
			give:   "auto-detect v1 by absence",
			config: EscrowOnChainConfigForTest{},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.config.IsV2())
		})
	}
}

// EscrowOnChainConfigForTest mirrors the config.EscrowOnChainConfig IsV2 logic
// to test it without importing the config package.
type EscrowOnChainConfigForTest struct {
	ContractVersion      string
	HubV2Address         string
	BeaconFactoryAddress string
}

func (c EscrowOnChainConfigForTest) IsV2() bool {
	if c.ContractVersion == "v2" {
		return true
	}
	if c.ContractVersion == "v1" {
		return false
	}
	return c.HubV2Address != "" || c.BeaconFactoryAddress != ""
}

func TestHubV2ClientWriteMethods(t *testing.T) {
	ctx := context.Background()
	caller := newMockCaller()
	client := NewHubV2Client(caller, common.HexToAddress("0x1000000000000000000000000000000000000001"), 84532)
	seller := common.HexToAddress("0x2000000000000000000000000000000000000002")
	token := common.HexToAddress("0x3000000000000000000000000000000000000003")
	member := common.HexToAddress("0x4000000000000000000000000000000000000004")
	refID := bytes32HubV2(0xab)

	tests := []struct {
		name     string
		method   string
		call     func() (string, error)
		wantArgs []interface{}
	}{
		{
			name:   "DirectSettle",
			method: MethodDirectSettle,
			call: func() (string, error) {
				return client.DirectSettle(ctx, seller, token, big.NewInt(10), refID)
			},
			wantArgs: []interface{}{seller, token, big.NewInt(10), refID},
		},
		{
			name:   "CompleteMilestone",
			method: MethodCompleteMilestone,
			call: func() (string, error) {
				return client.CompleteMilestone(ctx, big.NewInt(1), big.NewInt(2))
			},
			wantArgs: []interface{}{big.NewInt(1), big.NewInt(2)},
		},
		{
			name:   "ReleaseMilestone",
			method: MethodReleaseMilestone,
			call: func() (string, error) {
				return client.ReleaseMilestone(ctx, big.NewInt(3))
			},
			wantArgs: []interface{}{big.NewInt(3)},
		},
		{
			name:   "Deposit",
			method: MethodDeposit,
			call: func() (string, error) {
				return client.Deposit(ctx, big.NewInt(4))
			},
			wantArgs: []interface{}{big.NewInt(4)},
		},
		{
			name:   "Release",
			method: MethodRelease,
			call: func() (string, error) {
				return client.Release(ctx, big.NewInt(5))
			},
			wantArgs: []interface{}{big.NewInt(5)},
		},
		{
			name:   "Refund",
			method: MethodRefund,
			call: func() (string, error) {
				return client.Refund(ctx, big.NewInt(6))
			},
			wantArgs: []interface{}{big.NewInt(6)},
		},
		{
			name:   "Dispute",
			method: MethodDispute,
			call: func() (string, error) {
				return client.Dispute(ctx, big.NewInt(7))
			},
			wantArgs: []interface{}{big.NewInt(7)},
		},
		{
			name:   "ResolveDispute",
			method: MethodResolveDispute,
			call: func() (string, error) {
				return client.ResolveDispute(ctx, big.NewInt(8), big.NewInt(6), big.NewInt(2))
			},
			wantArgs: []interface{}{big.NewInt(8), big.NewInt(6), big.NewInt(2)},
		},
		{
			name:   "CreateMilestoneEscrow",
			method: MethodCreateMilestoneEscrow,
			call: func() (string, error) {
				_, tx, err := client.CreateMilestoneEscrow(
					ctx,
					seller,
					token,
					big.NewInt(30),
					[]*big.Int{big.NewInt(10), big.NewInt(20)},
					big.NewInt(99),
					refID,
				)
				return tx, err
			},
			wantArgs: []interface{}{seller, token, big.NewInt(30), []*big.Int{big.NewInt(10), big.NewInt(20)}, big.NewInt(99), refID},
		},
		{
			name:   "CreateTeamEscrow",
			method: MethodCreateTeamEscrow,
			call: func() (string, error) {
				_, tx, err := client.CreateTeamEscrow(
					ctx,
					[]common.Address{member},
					token,
					big.NewInt(40),
					[]*big.Int{big.NewInt(100)},
					big.NewInt(199),
					refID,
				)
				return tx, err
			},
			wantArgs: []interface{}{[]common.Address{member}, token, big.NewInt(40), []*big.Int{big.NewInt(100)}, big.NewInt(199), refID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := len(caller.writeCalls)
			tx, err := tt.call()
			require.NoError(t, err)
			assert.Equal(t, "0xmocktx", tx)
			require.Len(t, caller.writeCalls, before+1)
			call := caller.writeCalls[len(caller.writeCalls)-1]
			assert.Equal(t, tt.method, call.Method)
			assert.Equal(t, tt.wantArgs, call.Args)
		})
	}
}

func TestHubV2ClientCreateSimpleEscrowAndReadMethods(t *testing.T) {
	ctx := context.Background()
	caller := newMockCaller()
	caller.writeResult = &contract.ContractCallResult{
		Data:   []interface{}{big.NewInt(42)},
		TxHash: "0xcreated",
	}
	client := NewHubV2Client(caller, common.HexToAddress("0x1000000000000000000000000000000000000001"), 31337)

	dealID, txHash, err := client.CreateSimpleEscrow(
		ctx,
		common.HexToAddress("0x2000000000000000000000000000000000000002"),
		common.HexToAddress("0x3000000000000000000000000000000000000003"),
		big.NewInt(10),
		big.NewInt(20),
		bytes32HubV2(0x01),
	)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(42), dealID)
	assert.Equal(t, "0xcreated", txHash)
	require.Len(t, caller.writeCalls, 1)
	assert.Equal(t, MethodCreateSimpleEscrow, caller.writeCalls[0].Method)

	caller.readResult = &contract.ContractCallResult{Data: []interface{}{big.NewInt(43)}}
	nextID, err := client.NextDealID(ctx)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(43), nextID)
	require.Len(t, caller.readCalls, 1)
	assert.Equal(t, MethodNextDealID, caller.readCalls[0].Method)

	caller.readResult = &contract.ContractCallResult{}
	nextID, err = client.NextDealID(ctx)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0), nextID)
}

func TestHubV2ClientWrapsReadAndWriteErrors(t *testing.T) {
	ctx := context.Background()
	caller := newMockCaller()
	client := NewHubV2Client(caller, common.HexToAddress("0x1000000000000000000000000000000000000001"), 1)

	writeErr := errors.New("write failed")
	caller.writeErr = writeErr
	_, err := client.DirectSettle(
		ctx,
		common.Address{},
		common.Address{},
		big.NewInt(1),
		bytes32HubV2(0x01),
	)
	require.ErrorIs(t, err, writeErr)
	assert.Contains(t, err.Error(), MethodDirectSettle)

	readErr := errors.New("read failed")
	caller.writeErr = nil
	caller.readErr = readErr
	_, err = client.NextDealID(ctx)
	require.ErrorIs(t, err, readErr)
	assert.Contains(t, err.Error(), MethodNextDealID)
}

func TestParseDealV2Result(t *testing.T) {
	buyer := common.HexToAddress("0x1000000000000000000000000000000000000001")
	seller := common.HexToAddress("0x2000000000000000000000000000000000000002")
	token := common.HexToAddress("0x3000000000000000000000000000000000000003")
	settler := common.HexToAddress("0x4000000000000000000000000000000000000004")
	workHash := bytes32HubV2(0xaa)
	refID := bytes32HubV2(0xbb)

	deal, err := parseDealV2Result([]interface{}{
		struct {
			Buyer    common.Address
			Seller   common.Address
			Token    common.Address
			Amount   *big.Int
			Deadline *big.Int
			Status   uint8
			DealType uint8
			WorkHash [32]byte
			RefId    [32]byte
			Settler  common.Address
		}{
			Buyer:    buyer,
			Seller:   seller,
			Token:    token,
			Amount:   big.NewInt(10),
			Deadline: big.NewInt(20),
			Status:   uint8(DealStatusReleased),
			DealType: uint8(DealTypeTeam),
			WorkHash: workHash,
			RefId:    refID,
			Settler:  settler,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, buyer, deal.Buyer)
	assert.Equal(t, seller, deal.Seller)
	assert.Equal(t, token, deal.Token)
	assert.Equal(t, big.NewInt(10), deal.Amount)
	assert.Equal(t, big.NewInt(20), deal.Deadline)
	assert.Equal(t, DealStatusReleased, deal.Status)
	assert.Equal(t, DealTypeTeam, deal.DealType)
	assert.Equal(t, workHash, deal.WorkHash)
	assert.Equal(t, refID, deal.RefId)
	assert.Equal(t, settler, deal.Settler)

	_, err = parseDealV2Result(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty deal v2 result")

	_, err = parseDealV2Result([]interface{}{"bad"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected deal v2 result type")
}

func bytes32HubV2(last byte) [32]byte {
	var out [32]byte
	out[31] = last
	return out
}
