package smartaccount

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyShowRejectsInvalidOutputBeforeLoading(t *testing.T) {
	original := loadPolicyShowInfo
	called := false
	loadPolicyShowInfo = func(_ BootLoader) (policyShowInfo, func(), error) {
		called = true
		return policyShowInfo{}, nil, assert.AnError
	}
	t.Cleanup(func() { loadPolicyShowInfo = original })

	cmd := policyShowCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
	assert.Empty(t, out)
	assert.False(t, called)
}

func TestPolicyShowNoPolicyTableRunsCleanup(t *testing.T) {
	original := loadPolicyShowInfo
	cleanupCalled := false
	loadPolicyShowInfo = func(_ BootLoader) (policyShowInfo, func(), error) {
		return policyShowInfo{
			Account:   "0x1234abcd5678ef901234abcdef567890abcdef12",
			HasPolicy: false,
		}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { loadPolicyShowInfo = original })

	cmd := policyShowCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.NoError(t, err)
	assert.Contains(t, out, "Harness Policy")
	assert.Contains(t, out, "Status:")
	assert.Contains(t, out, "No policy set")
	assert.Contains(t, out, "Use 'lango account policy set' to configure limits.")
	assert.True(t, cleanupCalled)
}

func TestPolicyShowJSONRunsCleanup(t *testing.T) {
	original := loadPolicyShowInfo
	cleanupCalled := false
	loadPolicyShowInfo = func(_ BootLoader) (policyShowInfo, func(), error) {
		return policyShowInfo{
			Account:          "0x1234abcd5678ef901234abcdef567890abcdef12",
			HasPolicy:        true,
			MaxTxAmount:      "7",
			AutoApproveBelow: "3",
			AllowedTargets:   []string{"0x000000000000000000000000000000000000aaaa"},
			AllowedFunctions: []string{"0xa9059cbb"},
			RiskScore:        0.25,
		}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { loadPolicyShowInfo = original })

	cmd := policyShowCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "json")

	require.NoError(t, err)
	assert.True(t, cleanupCalled)

	var decoded policyShowInfo
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "0x1234abcd5678ef901234abcdef567890abcdef12", decoded.Account)
	assert.True(t, decoded.HasPolicy)
	assert.Equal(t, "7", decoded.MaxTxAmount)
	assert.Equal(t, []string{"0x000000000000000000000000000000000000aaaa"}, decoded.AllowedTargets)
	assert.Equal(t, []string{"0xa9059cbb"}, decoded.AllowedFunctions)
}

func TestPolicyShowPropagatesLoadErrorWithoutCleanup(t *testing.T) {
	original := loadPolicyShowInfo
	cleanupCalled := false
	loadPolicyShowInfo = func(_ BootLoader) (policyShowInfo, func(), error) {
		return policyShowInfo{}, func() { cleanupCalled = true }, fmt.Errorf("load policy failed")
	}
	t.Cleanup(func() { loadPolicyShowInfo = original })

	cmd := policyShowCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load policy failed")
	assert.Empty(t, out)
	assert.False(t, cleanupCalled)
}

func TestPolicySetPassesFlagsPrintsPartialResultAndRunsCleanup(t *testing.T) {
	original := updatePolicyLimits
	cleanupCalled := false
	updatePolicyLimits = func(_ BootLoader, maxTx, daily, monthly string) (policySetResult, func(), error) {
		assert.Equal(t, "100", maxTx)
		assert.Equal(t, "200", daily)
		assert.Empty(t, monthly)
		return policySetResult{
			Account:    "0x1234abcd5678ef901234abcdef567890abcdef12",
			DailyLimit: "200",
		}, func() { cleanupCalled = true }, nil
	}
	t.Cleanup(func() { updatePolicyLimits = original })

	cmd := policySetCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--max-tx", "100", "--daily", "200")

	require.NoError(t, err)
	assert.Contains(t, out, "Policy Updated")
	assert.Contains(t, out, "Account:")
	assert.Contains(t, out, "Daily Limit:")
	assert.Contains(t, out, "200")
	assert.NotContains(t, out, "Max Tx Amount:")
	assert.NotContains(t, out, "Monthly Limit:")
	assert.True(t, cleanupCalled)
}

func TestPolicySetRejectsPositionalArgsBeforeUpdating(t *testing.T) {
	original := updatePolicyLimits
	called := false
	updatePolicyLimits = func(_ BootLoader, _, _, _ string) (policySetResult, func(), error) {
		called = true
		return policySetResult{}, nil, nil
	}
	t.Cleanup(func() { updatePolicyLimits = original })

	cmd := policySetCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "unexpected", "--daily", "200")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown command "unexpected"`)
	assert.Empty(t, out)
	assert.False(t, called)
}

func TestPolicySetPropagatesUpdateErrorWithoutCleanup(t *testing.T) {
	original := updatePolicyLimits
	cleanupCalled := false
	updatePolicyLimits = func(_ BootLoader, _, _, _ string) (policySetResult, func(), error) {
		return policySetResult{}, func() { cleanupCalled = true }, fmt.Errorf("update failed")
	}
	t.Cleanup(func() { updatePolicyLimits = original })

	cmd := policySetCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--monthly", "300")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
	assert.Empty(t, out)
	assert.False(t, cleanupCalled)
}

func TestValueOrNA(t *testing.T) {
	assert.Equal(t, "n/a", valueOrNA(""))
	assert.Equal(t, "42", valueOrNA("42"))
}
