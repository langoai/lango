package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHubABI(t *testing.T) {
	t.Parallel()
	abi, err := ParseHubABI()
	require.NoError(t, err)
	require.NotNil(t, abi)

	expectedMethods := []string{"createDeal", "deposit", "submitWork", "release", "refund", "dispute", "resolveDispute", "getDeal", "nextDealId"}
	for _, m := range expectedMethods {
		_, ok := abi.Methods[m]
		assert.True(t, ok, "hub ABI missing method %q", m)
	}

	expectedEvents := []string{"DealCreated", "Deposited", "WorkSubmitted", "Released", "Refunded", "Disputed", "DealResolved"}
	for _, e := range expectedEvents {
		_, ok := abi.Events[e]
		assert.True(t, ok, "hub ABI missing event %q", e)
	}
}

func TestParseVaultABI(t *testing.T) {
	t.Parallel()
	abi, err := ParseVaultABI()
	require.NoError(t, err)
	require.NotNil(t, abi)

	expectedMethods := []string{"initialize", "deposit", "submitWork", "release", "refund", "dispute", "resolve"}
	for _, m := range expectedMethods {
		_, ok := abi.Methods[m]
		assert.True(t, ok, "vault ABI missing method %q", m)
	}

	expectedEvents := []string{"VaultInitialized", "Deposited", "WorkSubmitted", "Released", "Refunded", "Disputed", "VaultResolved"}
	for _, e := range expectedEvents {
		_, ok := abi.Events[e]
		assert.True(t, ok, "vault ABI missing event %q", e)
	}
}

func TestParseFactoryABI(t *testing.T) {
	t.Parallel()
	abi, err := ParseFactoryABI()
	require.NoError(t, err)
	require.NotNil(t, abi)

	expectedMethods := []string{"createVault", "getVault", "vaultCount"}
	for _, m := range expectedMethods {
		_, ok := abi.Methods[m]
		assert.True(t, ok, "factory ABI missing method %q", m)
	}

	_, ok := abi.Events["VaultCreated"]
	assert.True(t, ok, "factory ABI missing event VaultCreated")
}

func TestHubABIJSON_NotEmpty(t *testing.T) {
	t.Parallel()
	assert.NotEmpty(t, HubABIJSON())
}

func TestVaultABIJSON_NotEmpty(t *testing.T) {
	t.Parallel()
	assert.NotEmpty(t, VaultABIJSON())
}

func TestFactoryABIJSON_NotEmpty(t *testing.T) {
	t.Parallel()
	assert.NotEmpty(t, FactoryABIJSON())
}
