package smartaccount

import (
	"fmt"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartAccountPaymasterHelpers_WrapBootstrapErrors(t *testing.T) {
	bootErr := fmt.Errorf("boot unavailable")
	loader := func() (*bootstrap.Result, error) {
		return nil, bootErr
	}

	t.Run("status", func(t *testing.T) {
		info, cleanup, err := loadPaymasterStatus(loader)

		require.Error(t, err)
		assert.Equal(t, paymasterStatusInfo{}, info)
		assert.Nil(t, cleanup)
		assert.ErrorContains(t, err, "bootstrap: boot unavailable")
	})

	t.Run("approve", func(t *testing.T) {
		result, cleanup, err := executePaymasterApproval(loader, "1000.00")

		require.Error(t, err)
		assert.Equal(t, paymasterApproveResult{}, result)
		assert.Nil(t, cleanup)
		assert.ErrorContains(t, err, "bootstrap: boot unavailable")
	})
}

func TestSmartAccountPaymasterHelpers_ReturnDependencyInitializationErrors(t *testing.T) {
	loader := func() (*bootstrap.Result, error) {
		return &bootstrap.Result{Config: config.DefaultConfig()}, nil
	}

	t.Run("status", func(t *testing.T) {
		info, cleanup, err := loadPaymasterStatus(loader)

		require.Error(t, err)
		assert.Equal(t, paymasterStatusInfo{}, info)
		assert.Nil(t, cleanup)
		assert.ErrorContains(t, err, "smart account not enabled")
	})

	t.Run("approve", func(t *testing.T) {
		result, cleanup, err := executePaymasterApproval(loader, "1000.00")

		require.Error(t, err)
		assert.Equal(t, paymasterApproveResult{}, result)
		assert.Nil(t, cleanup)
		assert.ErrorContains(t, err, "smart account not enabled")
	})
}

func TestSmartAccountPaymasterCommands_TolerateNilCleanupOnSuccess(t *testing.T) {
	t.Run("status", func(t *testing.T) {
		original := loadPaymasterStatus
		loadPaymasterStatus = func(_ BootLoader) (paymasterStatusInfo, func(), error) {
			return paymasterStatusInfo{
				Enabled:          true,
				Provider:         "circle",
				Mode:             "rpc",
				TokenAddress:     "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PaymasterAddress: "0x31BE08D380A21fc740883c0BC434FcFc88740b58",
			}, nil, nil
		}
		t.Cleanup(func() { loadPaymasterStatus = original })

		cmd := paymasterStatusCmd(nil)
		out, err := executeSmartAccountPaymasterCmd(t, cmd)

		require.NoError(t, err)
		assert.Contains(t, out, "Paymaster Status")
		assert.Contains(t, out, "circle")
	})

	t.Run("approve", func(t *testing.T) {
		original := executePaymasterApproval
		executePaymasterApproval = func(_ BootLoader, amount string) (paymasterApproveResult, func(), error) {
			return paymasterApproveResult{
				Token:     "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				Paymaster: "0x31BE08D380A21fc740883c0BC434FcFc88740b58",
				Amount:    amount,
				TxHash:    "0xnilcleanup",
			}, nil, nil
		}
		t.Cleanup(func() { executePaymasterApproval = original })

		cmd := paymasterApproveCmd(nil)
		out, err := executeSmartAccountPaymasterCmd(t, cmd, "--amount", "12.34")

		require.NoError(t, err)
		assert.Contains(t, out, "Paymaster USDC Approval Submitted")
		assert.Contains(t, out, "12.34 USDC")
		assert.Contains(t, out, "0xnilcleanup")
	})
}
