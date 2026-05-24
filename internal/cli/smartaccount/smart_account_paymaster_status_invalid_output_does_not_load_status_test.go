package smartaccount

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartAccountPaymasterStatus_InvalidOutputDoesNotLoadStatus(t *testing.T) {
	original := loadPaymasterStatus
	var called bool
	loadPaymasterStatus = func(_ BootLoader) (paymasterStatusInfo, func(), error) {
		called = true
		return paymasterStatusInfo{}, nil, nil
	}
	t.Cleanup(func() { loadPaymasterStatus = original })

	cmd := paymasterStatusCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd, "--output", "yaml")

	require.Error(t, err)
	assert.False(t, called)
	assert.Empty(t, out)
	assert.ErrorContains(t, err, `unknown output format "yaml"`)
}

func TestSmartAccountPaymasterApprove_PropagatesApprovalError(t *testing.T) {
	original := executePaymasterApproval
	executePaymasterApproval = func(_ BootLoader, amount string) (paymasterApproveResult, func(), error) {
		assert.Equal(t, "42.50", amount)
		return paymasterApproveResult{}, nil, fmt.Errorf("approval rejected")
	}
	t.Cleanup(func() { executePaymasterApproval = original })

	cmd := paymasterApproveCmd(nil)
	out, err := executeSmartAccountPaymasterCmd(t, cmd, "--amount", "42.50")

	require.Error(t, err)
	assert.Empty(t, out)
	assert.ErrorContains(t, err, "approval rejected")
}

func TestSmartAccountPaymasterRootCommandWiresStatusAndApprove(t *testing.T) {
	cmd := paymasterCmd(nil)

	var names []string
	for _, child := range cmd.Commands() {
		names = append(names, child.Name())
	}

	assert.ElementsMatch(t, []string{"approve", "status"}, names)
	assert.Contains(t, cmd.Short, "paymaster")
}
