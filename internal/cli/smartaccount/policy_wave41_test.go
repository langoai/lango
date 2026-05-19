package smartaccount

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
)

func TestWave41PolicyShowTableFormatsPolicyFieldsAndRunsCleanup(t *testing.T) {
	cleanupCalled := false
	installWave41PolicyShowSeam(t, func(_ BootLoader) (policyShowInfo, func(), error) {
		return policyShowInfo{
			Account:          "0x1234abcd5678ef901234abcdef567890abcdef12",
			HasPolicy:        true,
			MaxTxAmount:      "100",
			DailyLimit:       "",
			MonthlyLimit:     "300",
			AutoApproveBelow: "",
			AllowedTargets: []string{
				"0x000000000000000000000000000000000000aaaa",
				"0x000000000000000000000000000000000000bbbb",
			},
			AllowedFunctions: []string{"0xa9059cbb", "0x095ea7b3", "0x23b872dd"},
			RiskScore:        0.75,
		}, func() { cleanupCalled = true }, nil
	})

	cmd := policyShowCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.NoError(t, err)
	assert.True(t, cleanupCalled)
	assert.Contains(t, out, "Harness Policy")
	assert.Contains(t, out, "Account:")
	assert.Contains(t, out, "0x1234abcd5678ef901234abcdef567890abcdef12")
	assert.Contains(t, out, "Max Tx Amount:")
	assert.Contains(t, out, "100")
	assert.Contains(t, out, "Daily Limit:")
	assert.Contains(t, out, "n/a")
	assert.Contains(t, out, "Monthly Limit:")
	assert.Contains(t, out, "300")
	assert.Contains(t, out, "Auto-Approve Below:")
	assert.Contains(t, out, "Required Risk Score:")
	assert.Contains(t, out, "0.75")
	assert.Contains(t, out, "Allowed Targets:")
	assert.Contains(t, out, "2 addresses")
	assert.Contains(t, out, "Allowed Functions:")
	assert.Contains(t, out, "3 selectors")
}

func TestWave41PolicyShowJSONFormatsPayloadAndRunsCleanup(t *testing.T) {
	cleanupCalled := false
	installWave41PolicyShowSeam(t, func(_ BootLoader) (policyShowInfo, func(), error) {
		return policyShowInfo{
			Account:          "0x1234abcd5678ef901234abcdef567890abcdef12",
			HasPolicy:        true,
			MaxTxAmount:      "100",
			DailyLimit:       "200",
			MonthlyLimit:     "300",
			AutoApproveBelow: "50",
			AllowedTargets:   []string{"0x000000000000000000000000000000000000aaaa"},
			AllowedFunctions: []string{"0xa9059cbb"},
			RiskScore:        0.5,
		}, func() { cleanupCalled = true }, nil
	})

	cmd := policyShowCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "json")

	require.NoError(t, err)
	assert.True(t, cleanupCalled)

	var decoded policyShowInfo
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "0x1234abcd5678ef901234abcdef567890abcdef12", decoded.Account)
	assert.True(t, decoded.HasPolicy)
	assert.Equal(t, "100", decoded.MaxTxAmount)
	assert.Equal(t, "200", decoded.DailyLimit)
	assert.Equal(t, "300", decoded.MonthlyLimit)
	assert.Equal(t, "50", decoded.AutoApproveBelow)
	assert.Equal(t, []string{"0x000000000000000000000000000000000000aaaa"}, decoded.AllowedTargets)
	assert.Equal(t, []string{"0xa9059cbb"}, decoded.AllowedFunctions)
	assert.Equal(t, 0.5, decoded.RiskScore)
}

func TestWave41PolicyShowRejectsInvalidOutputBeforeLoadingPolicy(t *testing.T) {
	called := false
	installWave41PolicyShowSeam(t, func(_ BootLoader) (policyShowInfo, func(), error) {
		called = true
		return policyShowInfo{}, nil, nil
	})

	cmd := policyShowCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--output", "xml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "xml"`)
	assert.Empty(t, out)
	assert.False(t, called)
}

func TestWave41PolicyShowPropagatesSeamErrorWithoutCleanup(t *testing.T) {
	cleanupCalled := false
	installWave41PolicyShowSeam(t, func(_ BootLoader) (policyShowInfo, func(), error) {
		return policyShowInfo{}, func() { cleanupCalled = true }, fmt.Errorf("policy show seam failed")
	})

	cmd := policyShowCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "policy show seam failed")
	assert.Empty(t, out)
	assert.False(t, cleanupCalled)
}

func TestWave41PolicySetTableFormatsReturnedLimitsAndRunsCleanup(t *testing.T) {
	cleanupCalled := false
	installWave41PolicySetSeam(t, func(_ BootLoader, maxTx, daily, monthly string) (policySetResult, func(), error) {
		assert.Equal(t, "100", maxTx)
		assert.Equal(t, "200", daily)
		assert.Equal(t, "300", monthly)
		return policySetResult{
			Account:      "0x1234abcd5678ef901234abcdef567890abcdef12",
			MaxTxAmount:  "100",
			DailyLimit:   "200",
			MonthlyLimit: "300",
		}, func() { cleanupCalled = true }, nil
	})

	cmd := policySetCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--max-tx", "100", "--daily", "200", "--monthly", "300")

	require.NoError(t, err)
	assert.True(t, cleanupCalled)
	assert.Contains(t, out, "Policy Updated")
	assert.Contains(t, out, "Account:")
	assert.Contains(t, out, "0x1234abcd5678ef901234abcdef567890abcdef12")
	assert.Contains(t, out, "Max Tx Amount:")
	assert.Contains(t, out, "100")
	assert.Contains(t, out, "Daily Limit:")
	assert.Contains(t, out, "200")
	assert.Contains(t, out, "Monthly Limit:")
	assert.Contains(t, out, "300")
}

func TestWave41PolicySetRunsCleanupWhenTableFlushFails(t *testing.T) {
	cleanupCalled := false
	installWave41PolicySetSeam(t, func(_ BootLoader, _, _, _ string) (policySetResult, func(), error) {
		return policySetResult{
			Account:      "0x1234abcd5678ef901234abcdef567890abcdef12",
			MaxTxAmount:  "100",
			DailyLimit:   "200",
			MonthlyLimit: "300",
		}, func() { cleanupCalled = true }, nil
	})

	cmd := policySetCmd(nil)
	cmd.SetOut(errorWriter{err: fmt.Errorf("writer failed")})
	cmd.SetArgs([]string{"--max-tx", "100"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "writer failed")
	assert.True(t, cleanupCalled)
}

func TestWave41PolicySetPropagatesNoLimitValidationFromSeam(t *testing.T) {
	cleanupCalled := false
	installWave41PolicySetSeam(t, func(_ BootLoader, maxTx, daily, monthly string) (policySetResult, func(), error) {
		assert.Empty(t, maxTx)
		assert.Empty(t, daily)
		assert.Empty(t, monthly)
		return policySetResult{}, func() { cleanupCalled = true }, fmt.Errorf("provide at least one policy limit (--max-tx, --daily, or --monthly)")
	})

	cmd := policySetCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provide at least one policy limit")
	assert.Empty(t, out)
	assert.False(t, cleanupCalled)
}

func TestWave41UpdatePolicyLimitsValidatesLimitsBeforeBootstrapping(t *testing.T) {
	tests := []struct {
		name      string
		maxTx     string
		daily     string
		monthly   string
		wantError string
	}{
		{
			name:      "missing limits",
			wantError: "provide at least one policy limit",
		},
		{
			name:      "invalid max tx",
			maxTx:     "1.5",
			wantError: `parse max-tx "1.5"`,
		},
		{
			name:      "invalid daily",
			daily:     "ten",
			wantError: `parse daily "ten"`,
		},
		{
			name:      "invalid monthly",
			monthly:   "12_000",
			wantError: `parse monthly "12_000"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootCalled := false
			result, cleanup, err := updatePolicyLimits(func() (*bootstrap.Result, error) {
				bootCalled = true
				return nil, fmt.Errorf("boot should not be called")
			}, tt.maxTx, tt.daily, tt.monthly)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
			assert.Equal(t, policySetResult{}, result)
			assert.Nil(t, cleanup)
			assert.False(t, bootCalled)
		})
	}
}

func TestWave41PolicySetPropagatesSeamErrorWithoutCleanup(t *testing.T) {
	cleanupCalled := false
	installWave41PolicySetSeam(t, func(_ BootLoader, maxTx, daily, monthly string) (policySetResult, func(), error) {
		assert.Equal(t, "100", maxTx)
		assert.Empty(t, daily)
		assert.Empty(t, monthly)
		return policySetResult{}, func() { cleanupCalled = true }, fmt.Errorf("policy set seam failed")
	})

	cmd := policySetCmd(nil)
	out, err := executeSmartAccountCmd(t, cmd, "--max-tx", "100")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "policy set seam failed")
	assert.Empty(t, out)
	assert.False(t, cleanupCalled)
}

func TestWave41ParsePolicyLimitAcceptsIntegersAndRejectsInvalidAmounts(t *testing.T) {
	got, err := parsePolicyLimit("max-tx", "1000000000000000000")
	require.NoError(t, err)
	assert.Equal(t, "1000000000000000000", got.String())

	tests := []struct {
		name      string
		wantField string
		wantValue string
	}{
		{
			name:      "max tx",
			wantField: "max-tx",
			wantValue: "1.5",
		},
		{
			name:      "daily",
			wantField: "daily",
			wantValue: "ten",
		},
		{
			name:      "monthly",
			wantField: "monthly",
			wantValue: "12_000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePolicyLimit(tt.wantField, tt.wantValue)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Contains(t, err.Error(), fmt.Sprintf("parse %s %q", tt.wantField, tt.wantValue))
			assert.Contains(t, err.Error(), "provide a wei amount (integer)")
		})
	}
}

func TestWave41ValueOrNAFormatsEmptyAndPresentValues(t *testing.T) {
	assert.Equal(t, "n/a", valueOrNA(""))
	assert.Equal(t, "0", valueOrNA("0"))
	assert.Equal(t, "1000000000000000000", valueOrNA("1000000000000000000"))
}

func installWave41PolicyShowSeam(
	t *testing.T,
	fn func(BootLoader) (policyShowInfo, func(), error),
) {
	t.Helper()
	original := loadPolicyShowInfo
	loadPolicyShowInfo = fn
	t.Cleanup(func() { loadPolicyShowInfo = original })
}

func installWave41PolicySetSeam(
	t *testing.T,
	fn func(BootLoader, string, string, string) (policySetResult, func(), error),
) {
	t.Helper()
	original := updatePolicyLimits
	updatePolicyLimits = fn
	t.Cleanup(func() { updatePolicyLimits = original })
}
