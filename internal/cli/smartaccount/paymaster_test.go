package smartaccount

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeSmartAccountPaymasterCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestSmartAccountPaymasterStatus_WritesTextToCommandWriter(t *testing.T) {
	original := loadPaymasterStatus
	loadPaymasterStatus = func(_ BootLoader) (paymasterStatusInfo, func(), error) {
		return paymasterStatusInfo{
			Enabled:          true,
			Provider:         "circle",
			Mode:             "permit",
			TokenAddress:     "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			PaymasterAddress: "0x31BE08D380A21fc740883c0BC434FcFc88740b58",
			ProviderType:     "circle-permit",
		}, func() {}, nil
	}
	t.Cleanup(func() { loadPaymasterStatus = original })

	cmd := paymasterStatusCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Paymaster Status")
	assert.Contains(t, out, "Provider:")
	assert.Contains(t, out, "circle")
}

func TestSmartAccountPaymasterStatus_WritesJSONToCommandWriter(t *testing.T) {
	original := loadPaymasterStatus
	loadPaymasterStatus = func(_ BootLoader) (paymasterStatusInfo, func(), error) {
		return paymasterStatusInfo{
			Enabled:          false,
			Provider:         "circle",
			Mode:             "permit",
			TokenAddress:     "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			PaymasterAddress: "0x31BE08D380A21fc740883c0BC434FcFc88740b58",
		}, func() {}, nil
	}
	t.Cleanup(func() { loadPaymasterStatus = original })

	cmd := paymasterStatusCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, false, decoded["enabled"])
	assert.Equal(t, "circle", decoded["provider"])
	assert.Equal(t, "permit", decoded["mode"])
}

func TestSmartAccountPaymasterApprove_WritesTextToCommandWriter(t *testing.T) {
	original := executePaymasterApproval
	executePaymasterApproval = func(_ BootLoader, amount string) (paymasterApproveResult, func(), error) {
		return paymasterApproveResult{
			Token:     "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			Paymaster: "0x31BE08D380A21fc740883c0BC434FcFc88740b58",
			Amount:    amount,
			TxHash:    "0xtxhash",
		}, func() {}, nil
	}
	t.Cleanup(func() { executePaymasterApproval = original })

	cmd := paymasterApproveCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd, "--amount", "1000.00")
	require.NoError(t, err)
	assert.Contains(t, out, "Paymaster USDC Approval Submitted")
	assert.Contains(t, out, "1000.00 USDC")
	assert.Contains(t, out, "0xtxhash")
}

func TestSmartAccountPaymasterApprove_WritesJSONToCommandWriter(t *testing.T) {
	original := executePaymasterApproval
	executePaymasterApproval = func(_ BootLoader, amount string) (paymasterApproveResult, func(), error) {
		return paymasterApproveResult{
			Token:     "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			Paymaster: "0x31BE08D380A21fc740883c0BC434FcFc88740b58",
			Amount:    amount,
			TxHash:    "0xtxhash",
		}, func() {}, nil
	}
	t.Cleanup(func() { executePaymasterApproval = original })

	cmd := paymasterApproveCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd, "--amount", "max", "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "max", decoded["amount"])
	assert.Equal(t, "0xtxhash", decoded["txHash"])
}
