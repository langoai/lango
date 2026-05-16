package smartaccount

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmartAccountPolicyShow_WritesTextToCommandWriter(t *testing.T) {
	original := loadPolicyShowInfo
	loadPolicyShowInfo = func(_ BootLoader) (policyShowInfo, func(), error) {
		return policyShowInfo{
			Account:          "0x1234abcd5678ef901234abcdef567890abcdef12",
			HasPolicy:        true,
			MaxTxAmount:      "5000000",
			DailyLimit:       "50000000",
			MonthlyLimit:     "500000000",
			AutoApproveBelow: "100000",
			AllowedTargets:   []string{"0xaaaa", "0xbbbb"},
			AllowedFunctions: []string{"0xa9059cbb"},
			RiskScore:        0.8,
		}, func() {}, nil
	}
	t.Cleanup(func() { loadPolicyShowInfo = original })

	cmd := policyShowCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Harness Policy")
	assert.Contains(t, out, "5000000")
	assert.Contains(t, out, "Allowed Targets:")
}

func TestSmartAccountPolicyShow_WritesJSONToCommandWriter(t *testing.T) {
	original := loadPolicyShowInfo
	loadPolicyShowInfo = func(_ BootLoader) (policyShowInfo, func(), error) {
		return policyShowInfo{
			Account:   "0x1234abcd5678ef901234abcdef567890abcdef12",
			HasPolicy: false,
		}, func() {}, nil
	}
	t.Cleanup(func() { loadPolicyShowInfo = original })

	cmd := policyShowCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "json")
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "0x1234abcd5678ef901234abcdef567890abcdef12", decoded["account"])
	assert.Equal(t, false, decoded["hasPolicy"])
}

func TestSmartAccountPolicySet_WritesTextToCommandWriter(t *testing.T) {
	original := updatePolicyLimits
	updatePolicyLimits = func(_ BootLoader, _, _, _ string) (policySetResult, func(), error) {
		return policySetResult{
			Account:      "0x1234abcd5678ef901234abcdef567890abcdef12",
			MaxTxAmount:  "5000000",
			DailyLimit:   "50000000",
			MonthlyLimit: "500000000",
		}, func() {}, nil
	}
	t.Cleanup(func() { updatePolicyLimits = original })

	cmd := policySetCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--max-tx", "5000000", "--daily", "50000000", "--monthly", "500000000")
	require.NoError(t, err)
	assert.Contains(t, out, "Policy Updated")
	assert.Contains(t, out, "500000000")
}
