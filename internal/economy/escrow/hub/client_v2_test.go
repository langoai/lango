package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
