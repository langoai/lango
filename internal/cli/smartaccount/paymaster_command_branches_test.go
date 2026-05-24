package smartaccount

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartAccountPaymasterStatus_TableIncludesOptionalFieldsAndRunsCleanup(t *testing.T) {
	original := loadPaymasterStatus
	cleanupCalled := false
	loadPaymasterStatus = func(_ BootLoader) (paymasterStatusInfo, func(), error) {
		return paymasterStatusInfo{
			Enabled:          true,
			Provider:         "pimlico",
			Mode:             "rpc",
			RPCURL:           "https://paymaster.example/rpc",
			TokenAddress:     "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			PaymasterAddress: "0x31BE08D380A21fc740883c0BC434FcFc88740b58",
			PolicyID:         "policy-123",
			ProviderType:     "pimlico",
		}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { loadPaymasterStatus = original })

	cmd := paymasterStatusCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "Provider Type:")
	assert.Contains(t, out, "pimlico")
	assert.Contains(t, out, "RPC URL:")
	assert.Contains(t, out, "https://paymaster.example/rpc")
	assert.Contains(t, out, "Policy ID:")
	assert.Contains(t, out, "policy-123")
	assert.True(t, cleanupCalled)
}

func TestSmartAccountPaymasterStatus_PropagatesLoadErrorWithoutCleanup(t *testing.T) {
	original := loadPaymasterStatus
	cleanupCalled := false
	loadPaymasterStatus = func(_ BootLoader) (paymasterStatusInfo, func(), error) {
		return paymasterStatusInfo{}, func() { cleanupCalled = true }, fmt.Errorf("status unavailable")
	}
	t.Cleanup(func() { loadPaymasterStatus = original })

	cmd := paymasterStatusCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd)

	require.Error(t, err)
	assert.Empty(t, out)
	assert.ErrorContains(t, err, "status unavailable")
	assert.False(t, cleanupCalled)
}

func TestSmartAccountPaymasterApprove_InvalidOutputDoesNotExecuteApproval(t *testing.T) {
	original := executePaymasterApproval
	called := false
	executePaymasterApproval = func(_ BootLoader, amount string) (paymasterApproveResult, func(), error) {
		called = true
		return paymasterApproveResult{}, nil, nil
	}
	t.Cleanup(func() { executePaymasterApproval = original })

	cmd := paymasterApproveCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd, "--output", "yaml")

	require.Error(t, err)
	assert.Empty(t, out)
	assert.ErrorContains(t, err, `unknown output format "yaml"`)
	assert.False(t, called)
}

func TestSmartAccountPaymasterApprove_TableUsesDefaultAmountAndRunsCleanup(t *testing.T) {
	original := executePaymasterApproval
	cleanupCalled := false
	executePaymasterApproval = func(_ BootLoader, amount string) (paymasterApproveResult, func(), error) {
		assert.Equal(t, "1000.00", amount)
		return paymasterApproveResult{
			Token:     "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			Paymaster: "0x31BE08D380A21fc740883c0BC434FcFc88740b58",
			Amount:    amount,
			TxHash:    "0xdefaultamount",
		}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { executePaymasterApproval = original })

	cmd := paymasterApproveCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "Paymaster USDC Approval Submitted")
	assert.Contains(t, out, "1000.00 USDC")
	assert.Contains(t, out, "0xdefaultamount")
	assert.True(t, cleanupCalled)
}

func TestSmartAccountPaymasterApprove_JSONPassesAmountAndRunsCleanup(t *testing.T) {
	original := executePaymasterApproval
	cleanupCalled := false
	executePaymasterApproval = func(_ BootLoader, amount string) (paymasterApproveResult, func(), error) {
		assert.Equal(t, "42.50", amount)
		return paymasterApproveResult{
			Token:     "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			Paymaster: "0x31BE08D380A21fc740883c0BC434FcFc88740b58",
			Amount:    amount,
			TxHash:    "0xjsonamount",
		}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { executePaymasterApproval = original })

	cmd := paymasterApproveCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd, "--amount", "42.50", "--output", "json")

	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "42.50", decoded["amount"])
	assert.Equal(t, "0xjsonamount", decoded["txHash"])
	assert.True(t, cleanupCalled)
}
