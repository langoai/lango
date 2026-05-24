package smartaccount

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	sa "github.com/langoai/lango/internal/smartaccount"
	sapolicy "github.com/langoai/lango/internal/smartaccount/policy"
)

func TestPolicyShowInfoUsesInjectedDepsAndFormatsExistingPolicy(t *testing.T) {
	account := common.HexToAddress("0x0000000000000000000000000000000000000abc")
	target := common.HexToAddress("0x0000000000000000000000000000000000000def")
	engine := sapolicy.New()
	engine.SetPolicy(account, &sapolicy.HarnessPolicy{
		MaxTxAmount:       big.NewInt(100),
		DailyLimit:        big.NewInt(200),
		MonthlyLimit:      big.NewInt(300),
		AutoApproveBelow:  big.NewInt(50),
		AllowedTargets:    []common.Address{target},
		AllowedFunctions:  []string{"0x12345678"},
		RequiredRiskScore: 0.75,
	})
	cleanupCalls := 0
	installSmartAccountPolicyDepsSeam(t, func(*bootstrap.Result) (*smartAccountDeps, error) {
		return &smartAccountDeps{
			manager:      fakePolicyAccountManager{info: &sa.AccountInfo{Address: account}},
			policyEngine: engine,
			cleanup:      func() { cleanupCalls++ },
		}, nil
	})

	result, cleanup, err := loadPolicyShowInfo(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})
	require.NoError(t, err)
	require.NotNil(t, cleanup)

	assert.Equal(t, account.Hex(), result.Account)
	assert.True(t, result.HasPolicy)
	assert.Equal(t, "100", result.MaxTxAmount)
	assert.Equal(t, "200", result.DailyLimit)
	assert.Equal(t, "300", result.MonthlyLimit)
	assert.Equal(t, "50", result.AutoApproveBelow)
	assert.Equal(t, []string{target.Hex()}, result.AllowedTargets)
	assert.Equal(t, []string{"0x12345678"}, result.AllowedFunctions)
	assert.Equal(t, 0.75, result.RiskScore)

	cleanup()
	assert.Equal(t, 1, cleanupCalls)
}

func TestPolicyShowInfoUsesInjectedDepsAndReportsNoPolicy(t *testing.T) {
	account := common.HexToAddress("0x0000000000000000000000000000000000000abc")
	cleanupCalls := 0
	installSmartAccountPolicyDepsSeam(t, func(*bootstrap.Result) (*smartAccountDeps, error) {
		return &smartAccountDeps{
			manager:      fakePolicyAccountManager{info: &sa.AccountInfo{Address: account}},
			policyEngine: sapolicy.New(),
			cleanup:      func() { cleanupCalls++ },
		}, nil
	})

	result, cleanup, err := loadPolicyShowInfo(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})
	require.NoError(t, err)
	require.NotNil(t, cleanup)

	assert.Equal(t, account.Hex(), result.Account)
	assert.False(t, result.HasPolicy)
	assert.Empty(t, result.MaxTxAmount)
	assert.Empty(t, result.DailyLimit)
	assert.Empty(t, result.MonthlyLimit)
	assert.Empty(t, result.AutoApproveBelow)
	assert.Empty(t, result.AllowedTargets)
	assert.Empty(t, result.AllowedFunctions)
	assert.Zero(t, result.RiskScore)

	cleanup()
	assert.Equal(t, 1, cleanupCalls)
}

func TestPolicyShowInfoCleansUpInjectedDepsWhenManagerInfoFails(t *testing.T) {
	infoErr := errors.New("manager info failed")
	cleanupCalls := 0
	installSmartAccountPolicyDepsSeam(t, func(*bootstrap.Result) (*smartAccountDeps, error) {
		return &smartAccountDeps{
			manager:      fakePolicyAccountManager{infoErr: infoErr},
			policyEngine: sapolicy.New(),
			cleanup:      func() { cleanupCalls++ },
		}, nil
	})

	result, cleanup, err := loadPolicyShowInfo(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get account info: manager info failed")
	assert.Equal(t, policyShowInfo{}, result)
	assert.Nil(t, cleanup)
	assert.Equal(t, 1, cleanupCalls)
}

func TestUpdatePolicyLimitsUsesInjectedDepsAndCreatesPolicy(t *testing.T) {
	account := common.HexToAddress("0x0000000000000000000000000000000000000abc")
	engine := sapolicy.New()
	cleanupCalls := 0
	installSmartAccountPolicyDepsSeam(t, func(*bootstrap.Result) (*smartAccountDeps, error) {
		return &smartAccountDeps{
			manager:      fakePolicyAccountManager{info: &sa.AccountInfo{Address: account}},
			policyEngine: engine,
			cleanup:      func() { cleanupCalls++ },
		}, nil
	})

	result, cleanup, err := updatePolicyLimits(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	}, "100", "200", "300")
	require.NoError(t, err)
	require.NotNil(t, cleanup)

	assert.Equal(t, account.Hex(), result.Account)
	assert.Equal(t, "100", result.MaxTxAmount)
	assert.Equal(t, "200", result.DailyLimit)
	assert.Equal(t, "300", result.MonthlyLimit)
	p, ok := engine.GetPolicy(account)
	require.True(t, ok)
	assert.Equal(t, "100", p.MaxTxAmount.String())
	assert.Equal(t, "200", p.DailyLimit.String())
	assert.Equal(t, "300", p.MonthlyLimit.String())

	cleanup()
	assert.Equal(t, 1, cleanupCalls)
}

func TestUpdatePolicyLimitsUsesInjectedDepsAndPreservesExistingPolicyFields(t *testing.T) {
	account := common.HexToAddress("0x0000000000000000000000000000000000000abc")
	target := common.HexToAddress("0x0000000000000000000000000000000000000def")
	engine := sapolicy.New()
	engine.SetPolicy(account, &sapolicy.HarnessPolicy{
		MaxTxAmount:       big.NewInt(100),
		DailyLimit:        big.NewInt(200),
		MonthlyLimit:      big.NewInt(300),
		AutoApproveBelow:  big.NewInt(25),
		AllowedTargets:    []common.Address{target},
		AllowedFunctions:  []string{"0xabcdef01"},
		RequiredRiskScore: 0.5,
	})
	installSmartAccountPolicyDepsSeam(t, func(*bootstrap.Result) (*smartAccountDeps, error) {
		return &smartAccountDeps{
			manager:      fakePolicyAccountManager{info: &sa.AccountInfo{Address: account}},
			policyEngine: engine,
			cleanup:      func() {},
		}, nil
	})

	result, cleanup, err := updatePolicyLimits(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	}, "", "250", "")
	require.NoError(t, err)
	require.NotNil(t, cleanup)
	cleanup()

	assert.Equal(t, "100", result.MaxTxAmount)
	assert.Equal(t, "250", result.DailyLimit)
	assert.Equal(t, "300", result.MonthlyLimit)
	p, ok := engine.GetPolicy(account)
	require.True(t, ok)
	assert.Equal(t, "100", p.MaxTxAmount.String())
	assert.Equal(t, "250", p.DailyLimit.String())
	assert.Equal(t, "300", p.MonthlyLimit.String())
	assert.Equal(t, "25", p.AutoApproveBelow.String())
	assert.Equal(t, []common.Address{target}, p.AllowedTargets)
	assert.Equal(t, []string{"0xabcdef01"}, p.AllowedFunctions)
	assert.Equal(t, 0.5, p.RequiredRiskScore)
}

func TestUpdatePolicyLimitsCleansUpInjectedDepsWhenManagerInfoFails(t *testing.T) {
	infoErr := errors.New("manager info failed")
	cleanupCalls := 0
	installSmartAccountPolicyDepsSeam(t, func(*bootstrap.Result) (*smartAccountDeps, error) {
		return &smartAccountDeps{
			manager:      fakePolicyAccountManager{infoErr: infoErr},
			policyEngine: sapolicy.New(),
			cleanup:      func() { cleanupCalls++ },
		}, nil
	})

	result, cleanup, err := updatePolicyLimits(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	}, "100", "", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "get account info: manager info failed")
	assert.Equal(t, policySetResult{}, result)
	assert.Nil(t, cleanup)
	assert.Equal(t, 1, cleanupCalls)
}

func installSmartAccountPolicyDepsSeam(
	t *testing.T,
	fn func(*bootstrap.Result) (*smartAccountDeps, error),
) {
	t.Helper()
	original := initSmartAccountPolicyDeps
	initSmartAccountPolicyDeps = fn
	t.Cleanup(func() { initSmartAccountPolicyDeps = original })
}

type fakePolicyAccountManager struct {
	info    *sa.AccountInfo
	infoErr error
}

func (m fakePolicyAccountManager) GetOrDeploy(ctx context.Context) (*sa.AccountInfo, error) {
	return nil, errors.New("unexpected GetOrDeploy call")
}

func (m fakePolicyAccountManager) Info(context.Context) (*sa.AccountInfo, error) {
	if m.infoErr != nil {
		return nil, m.infoErr
	}
	return m.info, nil
}

func (fakePolicyAccountManager) InstallModule(
	context.Context,
	sa.ModuleType,
	common.Address,
	[]byte,
) (string, error) {
	return "", errors.New("unexpected InstallModule call")
}

func (fakePolicyAccountManager) UninstallModule(
	context.Context,
	sa.ModuleType,
	common.Address,
	[]byte,
) (string, error) {
	return "", errors.New("unexpected UninstallModule call")
}

func (fakePolicyAccountManager) Execute(context.Context, []sa.ContractCall) (string, error) {
	return "", errors.New("unexpected Execute call")
}
