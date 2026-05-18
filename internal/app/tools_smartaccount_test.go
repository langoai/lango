package app

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/economy/budget"
	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/smartaccount/module"
	"github.com/langoai/lango/internal/smartaccount/paymaster"
	"github.com/langoai/lango/internal/smartaccount/policy"
	sasession "github.com/langoai/lango/internal/smartaccount/session"
)

func TestBuildSmartAccountTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := buildSmartAccountTools(nil)

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "session_key_create requires targets",
			tool:    "session_key_create",
			params:  map[string]interface{}{"duration": "1h"},
			wantErr: "missing targets parameter",
		},
		{
			name:    "session_key_create requires duration",
			tool:    "session_key_create",
			params:  map[string]interface{}{"targets": []interface{}{"0x123"}},
			wantErr: "missing duration parameter",
		},
		{
			name:    "session_key_revoke requires session_id",
			tool:    "session_key_revoke",
			params:  map[string]interface{}{},
			wantErr: "missing session_id parameter",
		},
		{
			name:    "session_execute requires session_id",
			tool:    "session_execute",
			params:  map[string]interface{}{"target": "0x123"},
			wantErr: "missing session_id parameter",
		},
		{
			name:    "session_execute requires target",
			tool:    "session_execute",
			params:  map[string]interface{}{"session_id": "sess-1"},
			wantErr: "missing target parameter",
		},
		{
			name:    "policy_check requires target",
			tool:    "policy_check",
			params:  map[string]interface{}{},
			wantErr: "missing target parameter",
		},
		{
			name:    "module_install requires module_type",
			tool:    "module_install",
			params:  map[string]interface{}{"address": "0x123"},
			wantErr: "missing module_type parameter",
		},
		{
			name:    "module_install requires address",
			tool:    "module_install",
			params:  map[string]interface{}{"module_type": float64(1)},
			wantErr: "missing address parameter",
		},
		{
			name:    "module_uninstall requires module_type",
			tool:    "module_uninstall",
			params:  map[string]interface{}{"address": "0x123"},
			wantErr: "missing module_type parameter",
		},
		{
			name:    "module_uninstall requires address",
			tool:    "module_uninstall",
			params:  map[string]interface{}{"module_type": float64(1)},
			wantErr: "missing address parameter",
		},
		{
			name:    "paymaster_approve requires token_address",
			tool:    "paymaster_approve",
			params:  map[string]interface{}{"paymaster_address": "0x123", "amount": "1.00"},
			wantErr: "missing token_address parameter",
		},
		{
			name:    "paymaster_approve requires paymaster_address",
			tool:    "paymaster_approve",
			params:  map[string]interface{}{"token_address": "0x123", "amount": "1.00"},
			wantErr: "missing paymaster_address parameter",
		},
		{
			name:    "paymaster_approve requires amount",
			tool:    "paymaster_approve",
			params:  map[string]interface{}{"token_address": "0x123", "paymaster_address": "0xabc"},
			wantErr: "missing amount parameter",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := findSmartAccountTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestBuildSmartAccountTools_MetadataInventory(t *testing.T) {
	t.Parallel()

	tools := buildSmartAccountTools(nil)
	require.Len(t, tools, 12)

	expected := map[string]agent.SafetyLevel{
		"smart_account_deploy": agent.SafetyLevelDangerous,
		"smart_account_info":   agent.SafetyLevelSafe,
		"session_key_create":   agent.SafetyLevelDangerous,
		"session_key_list":     agent.SafetyLevelSafe,
		"session_key_revoke":   agent.SafetyLevelDangerous,
		"session_execute":      agent.SafetyLevelDangerous,
		"policy_check":         agent.SafetyLevelSafe,
		"module_install":       agent.SafetyLevelDangerous,
		"module_uninstall":     agent.SafetyLevelDangerous,
		"spending_status":      agent.SafetyLevelSafe,
		"paymaster_status":     agent.SafetyLevelSafe,
		"paymaster_approve":    agent.SafetyLevelDangerous,
	}

	seen := make(map[string]bool, len(tools))
	for _, tool := range tools {
		wantSafety, ok := expected[tool.Name]
		require.True(t, ok, "unexpected tool %q", tool.Name)
		assert.Equal(t, wantSafety, tool.SafetyLevel)
		assert.Equal(t, "smartaccount", tool.Capability.Category)
		assert.Equal(t, agent.ExposureDeferred, tool.Capability.Exposure)
		assert.Equal(t, agent.ActivityManage, tool.Capability.Activity)
		assert.Contains(t, tool.Capability.RequiredCapabilities, "payment")
		assert.NotEmpty(t, tool.Description)
		assert.NotNil(t, tool.Parameters)
		seen[tool.Name] = true
	}
	assert.Len(t, seen, len(expected))
}

func TestSmartAccountDeployAndInfoTools_ReturnAccountMetadata(t *testing.T) {
	t.Parallel()

	accountInfo := &sa.AccountInfo{
		Address:      common.HexToAddress("0x1000000000000000000000000000000000000001"),
		IsDeployed:   true,
		OwnerAddress: common.HexToAddress("0x2000000000000000000000000000000000000002"),
		ChainID:      84532,
		EntryPoint:   common.HexToAddress("0x3000000000000000000000000000000000000003"),
		Modules: []sa.ModuleInfo{
			{
				Address: common.HexToAddress("0x4000000000000000000000000000000000000004"),
				Type:    sa.ModuleTypeValidator,
				Name:    "SessionValidator",
			},
		},
	}
	manager := &fakeSmartAccountManager{info: accountInfo}
	sac := newSmartAccountTestComponents(manager)
	tools := buildSmartAccountTools(sac)

	for _, name := range []string{"smart_account_deploy", "smart_account_info"} {
		name := name
		t.Run(name, func(t *testing.T) {
			got, err := findSmartAccountTool(t, tools, name).Handler(context.Background(), map[string]interface{}{})
			require.NoError(t, err)

			result := requireMap(t, got)
			assert.Equal(t, accountInfo.Address.Hex(), result["address"])
			assert.Equal(t, true, result["isDeployed"])
			assert.Equal(t, accountInfo.OwnerAddress.Hex(), result["ownerAddress"])
			assert.Equal(t, int64(84532), result["chainId"])
			assert.Equal(t, accountInfo.EntryPoint.Hex(), result["entryPoint"])

			modules := requireMapSlice(t, result["modules"])
			require.Len(t, modules, 1)
			assert.Equal(t, "SessionValidator", modules[0]["name"])
			assert.Equal(t, "validator", modules[0]["type"])
			assert.Equal(t, accountInfo.Modules[0].Address.Hex(), modules[0]["address"])
		})
	}
}

func TestSmartAccountDeployTool_ManagerError(t *testing.T) {
	t.Parallel()

	manager := &fakeSmartAccountManager{deployErr: errors.New("bundler unavailable")}
	sac := newSmartAccountTestComponents(manager)

	got, err := smartAccountDeployTool(sac).Handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "deploy smart account")
	assert.Contains(t, err.Error(), "bundler unavailable")
}

func TestSessionKeyCreateAndListTools_Success(t *testing.T) {
	t.Parallel()

	sac := newSmartAccountTestComponents(&fakeSmartAccountManager{})
	tools := buildSmartAccountTools(sac)

	created, err := findSmartAccountTool(t, tools, "session_key_create").Handler(context.Background(), map[string]interface{}{
		"targets":     []interface{}{"0x5000000000000000000000000000000000000005"},
		"functions":   []interface{}{"0xa9059cbb"},
		"spend_limit": "10.50",
		"duration":    "30m",
	})
	require.NoError(t, err)
	createdMap := requireMap(t, created)
	assert.NotEmpty(t, createdMap["sessionId"])
	assert.NotEmpty(t, createdMap["address"])
	assert.Empty(t, createdMap["parentId"])
	assert.Equal(t, 1, createdMap["targets"])
	assert.Equal(t, 1, createdMap["functions"])
	assert.NotEmpty(t, createdMap["expiresAt"])

	listed, err := findSmartAccountTool(t, tools, "session_key_list").Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	listedMap := requireMap(t, listed)
	assert.Equal(t, 1, listedMap["total"])

	sessions := requireMapSlice(t, listedMap["sessions"])
	require.Len(t, sessions, 1)
	assert.Equal(t, createdMap["sessionId"], sessions[0]["sessionId"])
	assert.Equal(t, "active", sessions[0]["status"])
	assert.Empty(t, sessions[0]["parentId"])
	assert.NotEmpty(t, sessions[0]["createdAt"])
	assert.NotEmpty(t, sessions[0]["expiresAt"])
}

func TestPolicyCheckTool_ReturnsAllowedAndDenied(t *testing.T) {
	t.Parallel()

	target := common.HexToAddress("0x6000000000000000000000000000000000000006")
	engine := policy.New()
	engine.SetPolicy(target, &policy.HarnessPolicy{
		MaxTxAmount:      big.NewInt(100),
		AllowedTargets:   []common.Address{target},
		AllowedFunctions: []string{"transfer(address,uint256)"},
	})
	sac := newSmartAccountTestComponents(&fakeSmartAccountManager{})
	sac.policyEngine = engine
	tool := policyCheckTool(sac)

	allowed, err := tool.Handler(context.Background(), map[string]interface{}{
		"target":       target.Hex(),
		"value":        "99",
		"function_sig": "transfer(address,uint256)",
	})
	require.NoError(t, err)
	allowedMap := requireMap(t, allowed)
	assert.Equal(t, true, allowedMap["allowed"])
	assert.Equal(t, target.Hex(), allowedMap["target"])

	denied, err := tool.Handler(context.Background(), map[string]interface{}{
		"target":       target.Hex(),
		"value":        "101",
		"function_sig": "transfer(address,uint256)",
	})
	require.NoError(t, err)
	deniedMap := requireMap(t, denied)
	assert.Equal(t, false, deniedMap["allowed"])
	assert.Contains(t, deniedMap["reason"], "exceeds max")
}

func TestModuleInstallAndUninstallTools_SuccessAndErrors(t *testing.T) {
	t.Parallel()

	moduleAddr := common.HexToAddress("0x7000000000000000000000000000000000000007")
	manager := &fakeSmartAccountManager{
		installHash:   "0xinstall",
		uninstallHash: "0xuninstall",
	}
	sac := newSmartAccountTestComponents(manager)
	tools := buildSmartAccountTools(sac)

	installed, err := findSmartAccountTool(t, tools, "module_install").Handler(context.Background(), map[string]interface{}{
		"module_type": float64(1),
		"address":     moduleAddr.Hex(),
		"init_data":   "0x1234",
	})
	require.NoError(t, err)
	installedMap := requireMap(t, installed)
	assert.Equal(t, "0xinstall", installedMap["txHash"])
	assert.Equal(t, "validator", installedMap["moduleType"])
	assert.Equal(t, moduleAddr.Hex(), installedMap["address"])
	assert.Equal(t, "installed", installedMap["status"])
	assert.Equal(t, sa.ModuleTypeValidator, manager.lastInstallType)
	assert.Equal(t, moduleAddr, manager.lastInstallAddr)
	assert.Equal(t, []byte{0x12, 0x34}, manager.lastInstallData)

	uninstalled, err := findSmartAccountTool(t, tools, "module_uninstall").Handler(context.Background(), map[string]interface{}{
		"module_type": float64(4),
		"address":     moduleAddr.Hex(),
	})
	require.NoError(t, err)
	uninstalledMap := requireMap(t, uninstalled)
	assert.Equal(t, "0xuninstall", uninstalledMap["txHash"])
	assert.Equal(t, "hook", uninstalledMap["moduleType"])
	assert.Equal(t, moduleAddr.Hex(), uninstalledMap["address"])
	assert.Equal(t, "uninstalled", uninstalledMap["status"])
	assert.Equal(t, sa.ModuleTypeHook, manager.lastUninstallType)
	assert.Equal(t, moduleAddr, manager.lastUninstallAddr)

	manager.installErr = errors.New("install failed")
	got, err := findSmartAccountTool(t, tools, "module_install").Handler(context.Background(), map[string]interface{}{
		"module_type": float64(1),
		"address":     moduleAddr.Hex(),
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.EqualError(t, err, "install failed")

	manager.uninstallErr = errors.New("uninstall failed")
	got, err = findSmartAccountTool(t, tools, "module_uninstall").Handler(context.Background(), map[string]interface{}{
		"module_type": float64(1),
		"address":     moduleAddr.Hex(),
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.EqualError(t, err, "uninstall failed")
}

func TestSpendingStatusTool_ReturnsTrackedSpendAndModules(t *testing.T) {
	t.Parallel()

	sac := newSmartAccountTestComponents(&fakeSmartAccountManager{})
	sessionID := "session-1"
	sac.onChainTracker.Record(sessionID, big.NewInt(12345))
	require.NoError(t, sac.moduleRegistry.Register(&module.ModuleDescriptor{
		Name:    "SpendHook",
		Address: common.HexToAddress("0x8000000000000000000000000000000000000008"),
		Type:    sa.ModuleTypeHook,
		Version: "1.2.3",
	}))

	got, err := spendingStatusTool(sac).Handler(context.Background(), map[string]interface{}{"session_id": sessionID})
	require.NoError(t, err)
	result := requireMap(t, got)
	assert.Equal(t, sessionID, result["sessionId"])
	assert.Equal(t, "12345", result["onChainSpent"])

	modules := requireMapSlice(t, result["registeredModules"])
	require.Len(t, modules, 1)
	assert.Equal(t, "SpendHook", modules[0]["name"])
	assert.Equal(t, "hook", modules[0]["type"])
	assert.Equal(t, "1.2.3", modules[0]["version"])
}

func TestPaymasterStatusTool_ReturnsProviderState(t *testing.T) {
	t.Parallel()

	disabled, err := paymasterStatusTool(newSmartAccountTestComponents(&fakeSmartAccountManager{})).
		Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	disabledMap := requireMap(t, disabled)
	assert.Equal(t, false, disabledMap["enabled"])
	assert.Equal(t, "none", disabledMap["provider"])

	sac := newSmartAccountTestComponents(&fakeSmartAccountManager{})
	sac.paymasterProvider = fakePaymasterProvider{providerType: "circle-test"}
	enabled, err := paymasterStatusTool(sac).Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	enabledMap := requireMap(t, enabled)
	assert.Equal(t, true, enabledMap["enabled"])
	assert.Equal(t, "circle-test", enabledMap["provider"])
}

func TestPaymasterApproveTool_SuccessMaxAndErrors(t *testing.T) {
	t.Parallel()

	tokenAddr := common.HexToAddress("0x9000000000000000000000000000000000000009")
	paymasterAddr := common.HexToAddress("0xa00000000000000000000000000000000000000a")
	manager := &fakeSmartAccountManager{executeHash: "0xapprove"}
	sac := newSmartAccountTestComponents(manager)
	tool := paymasterApproveTool(sac)

	got, err := tool.Handler(context.Background(), map[string]interface{}{
		"token_address":     tokenAddr.Hex(),
		"paymaster_address": paymasterAddr.Hex(),
		"amount":            "max",
	})
	require.NoError(t, err)
	result := requireMap(t, got)
	assert.Equal(t, "0xapprove", result["txHash"])
	assert.Equal(t, tokenAddr.Hex(), result["token"])
	assert.Equal(t, paymasterAddr.Hex(), result["paymaster"])
	assert.Equal(t, "max", result["amount"])
	assert.Equal(t, "approved", result["status"])
	require.Len(t, manager.lastExecuteCalls, 1)
	assert.Equal(t, tokenAddr, manager.lastExecuteCalls[0].Target)
	assert.Equal(t, "approve(address,uint256)", manager.lastExecuteCalls[0].FunctionSig)
	assert.NotEmpty(t, manager.lastExecuteCalls[0].Data)

	got, err = tool.Handler(context.Background(), map[string]interface{}{
		"token_address":     tokenAddr.Hex(),
		"paymaster_address": paymasterAddr.Hex(),
		"amount":            "not-usdc",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "parse amount")

	manager.executeErr = errors.New("execution rejected")
	got, err = tool.Handler(context.Background(), map[string]interface{}{
		"token_address":     tokenAddr.Hex(),
		"paymaster_address": paymasterAddr.Hex(),
		"amount":            "1.00",
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "approve USDC")
	assert.Contains(t, err.Error(), "execution rejected")
}

func findSmartAccountTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}

func newSmartAccountTestComponents(manager sa.AccountManager) *smartAccountComponents {
	return &smartAccountComponents{
		manager:        manager,
		sessionManager: sasession.NewManager(sasession.NewMemoryStore(), sasession.WithMaxDuration(time.Hour)),
		policyEngine:   policy.New(),
		moduleRegistry: module.NewRegistry(),
		onChainTracker: budget.NewOnChainTracker(),
	}
}

func requireMap(t *testing.T, got interface{}) map[string]interface{} {
	t.Helper()

	result, ok := got.(map[string]interface{})
	require.Truef(t, ok, "expected map[string]interface{}, got %T", got)
	return result
}

func requireMapSlice(t *testing.T, got interface{}) []map[string]interface{} {
	t.Helper()

	result, ok := got.([]map[string]interface{})
	require.Truef(t, ok, "expected []map[string]interface{}, got %T", got)
	return result
}

type fakeSmartAccountManager struct {
	info          *sa.AccountInfo
	deployErr     error
	infoErr       error
	installErr    error
	uninstallErr  error
	executeErr    error
	installHash   string
	uninstallHash string
	executeHash   string

	lastInstallType   sa.ModuleType
	lastInstallAddr   common.Address
	lastInstallData   []byte
	lastUninstallType sa.ModuleType
	lastUninstallAddr common.Address
	lastExecuteCalls  []sa.ContractCall
}

func (m *fakeSmartAccountManager) GetOrDeploy(context.Context) (*sa.AccountInfo, error) {
	if m.deployErr != nil {
		return nil, m.deployErr
	}
	return m.info, nil
}

func (m *fakeSmartAccountManager) Info(context.Context) (*sa.AccountInfo, error) {
	if m.infoErr != nil {
		return nil, m.infoErr
	}
	return m.info, nil
}

func (m *fakeSmartAccountManager) InstallModule(_ context.Context, moduleType sa.ModuleType, addr common.Address, initData []byte) (string, error) {
	if m.installErr != nil {
		return "", m.installErr
	}
	m.lastInstallType = moduleType
	m.lastInstallAddr = addr
	m.lastInstallData = append([]byte(nil), initData...)
	return m.installHash, nil
}

func (m *fakeSmartAccountManager) UninstallModule(_ context.Context, moduleType sa.ModuleType, addr common.Address, _ []byte) (string, error) {
	if m.uninstallErr != nil {
		return "", m.uninstallErr
	}
	m.lastUninstallType = moduleType
	m.lastUninstallAddr = addr
	return m.uninstallHash, nil
}

func (m *fakeSmartAccountManager) Execute(_ context.Context, calls []sa.ContractCall) (string, error) {
	if m.executeErr != nil {
		return "", m.executeErr
	}
	m.lastExecuteCalls = append([]sa.ContractCall(nil), calls...)
	return m.executeHash, nil
}

type fakePaymasterProvider struct {
	providerType string
}

func (p fakePaymasterProvider) SponsorUserOp(context.Context, *paymaster.SponsorRequest) (*paymaster.SponsorResult, error) {
	return &paymaster.SponsorResult{PaymasterAndData: []byte{0x01}}, nil
}

func (p fakePaymasterProvider) Type() string { return p.providerType }
