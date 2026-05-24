package settings

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
)

func TestAgentFormWiresProviderModelBranchesAndReloads(t *testing.T) {
	modelServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/models", r.URL.Path)
		fmt.Fprint(w, `{"object":"list","data":[{"id":"z-model"},{"id":"a-model"}]}`)
	}))
	defer modelServer.Close()

	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "primary"
	cfg.Agent.Model = "current-agent-model"
	cfg.Agent.MaxTokens = 2048
	cfg.Agent.Temperature = 0.7
	cfg.Agent.PromptsDir = "/tmp/prompts"
	cfg.Agent.FallbackProvider = "fallback"
	cfg.Agent.FallbackModel = "current-fallback-model"
	cfg.Agent.RequestTimeout = 45 * time.Second
	cfg.Agent.ToolTimeout = 12 * time.Second
	cfg.Agent.AutoExtendTimeout = true
	cfg.Agent.MaxRequestTimeout = 2 * time.Minute
	cfg.Providers = map[string]config.ProviderConfig{
		"zeta": {
			Type: types.ProviderAnthropic,
		},
		"primary": {
			Type:    types.ProviderOpenAI,
			APIKey:  "test-key",
			BaseURL: modelServer.URL + "/v1",
		},
		"fallback": {
			Type:    types.ProviderOpenAI,
			APIKey:  "test-key",
			BaseURL: modelServer.URL + "/v1",
		},
	}

	form := NewAgentForm(cfg)

	assert.Equal(t, "Agent Configuration", form.Title)
	assert.Equal(t, []string{
		"provider",
		"model",
		"maxtokens",
		"temp",
		"prompts_dir",
		"fallback_provider",
		"fallback_model",
		"request_timeout",
		"tool_timeout",
		"auto_extend_timeout",
		"max_request_timeout",
	}, formKeys(form))

	provider := fieldByKey(form, "provider")
	require.NotNil(t, provider)
	assert.Equal(t, tuicore.InputSelect, provider.Type)
	assert.Equal(t, "primary", provider.Value)
	assert.Equal(t, []string{"fallback", "primary", "zeta"}, provider.Options)

	model := fieldByKey(form, "model")
	require.NotNil(t, model)
	assert.Equal(t, tuicore.InputSearchSelect, model.Type)
	assert.Equal(t, "current-agent-model", model.Value)
	assert.Equal(t, []string{"current-agent-model", "a-model", "z-model"}, model.Options)
	assert.Empty(t, model.Placeholder)
	assert.Contains(t, model.Description, "Fetched 3 models")

	maxTokens := fieldByKey(form, "maxtokens")
	require.NotNil(t, maxTokens)
	assert.Equal(t, "2048", maxTokens.Value)
	require.NoError(t, maxTokens.Validate("0"))
	assert.EqualError(t, maxTokens.Validate("not-int"), "must be integer")

	temp := fieldByKey(form, "temp")
	require.NotNil(t, temp)
	assert.Equal(t, "0.7", temp.Value)
	require.NoError(t, temp.Validate("2.0"))
	assert.EqualError(t, temp.Validate("-0.1"), "must be between 0.0 and 2.0")
	assert.EqualError(t, temp.Validate("hot"), "must be a number")

	assert.Equal(t, "/tmp/prompts", fieldByKey(form, "prompts_dir").Value)
	assert.Equal(t, "45s", fieldByKey(form, "request_timeout").Value)
	assert.Equal(t, "12s", fieldByKey(form, "tool_timeout").Value)
	assert.True(t, fieldByKey(form, "auto_extend_timeout").Checked)
	assert.Equal(t, "2m0s", fieldByKey(form, "max_request_timeout").Value)

	fallbackProvider := fieldByKey(form, "fallback_provider")
	require.NotNil(t, fallbackProvider)
	assert.Equal(t, tuicore.InputSelect, fallbackProvider.Type)
	assert.Equal(t, "fallback", fallbackProvider.Value)
	assert.Equal(t, []string{"", "fallback", "primary", "zeta"}, fallbackProvider.Options)

	fallbackModel := fieldByKey(form, "fallback_model")
	require.NotNil(t, fallbackModel)
	assert.Equal(t, tuicore.InputSearchSelect, fallbackModel.Type)
	assert.Equal(t, "current-fallback-model", fallbackModel.Value)
	assert.Equal(t, []string{"", "current-fallback-model", "a-model", "z-model"}, fallbackModel.Options)
	assert.Empty(t, fallbackModel.Placeholder)

	cmd := provider.OnChange("fallback")
	require.NotNil(t, cmd)
	assert.True(t, model.Loading)
	msg, ok := cmd().(tuicore.FieldOptionsLoadedMsg)
	require.True(t, ok)
	assert.Equal(t, "model", msg.FieldKey)
	assert.Equal(t, "fallback", msg.ProviderID)
	require.NoError(t, msg.Err)
	assert.Equal(t, []string{"a-model", "z-model"}, msg.Options)

	assert.Nil(t, fallbackProvider.OnChange(""))
	fallbackCmd := fallbackProvider.OnChange("primary")
	require.NotNil(t, fallbackCmd)
	assert.True(t, fallbackModel.Loading)
	fallbackMsg, ok := fallbackCmd().(tuicore.FieldOptionsLoadedMsg)
	require.True(t, ok)
	assert.Equal(t, "fallback_model", fallbackMsg.FieldKey)
	assert.Equal(t, "primary", fallbackMsg.ProviderID)
	require.NoError(t, fallbackMsg.Err)
}

func TestAgentFormFallsBackToManualModelFields(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "missing-key"
	cfg.Agent.Model = "manual-agent-model"
	cfg.Agent.FallbackProvider = "unsupported"
	cfg.Agent.FallbackModel = "manual-fallback-model"
	cfg.Providers = map[string]config.ProviderConfig{
		"missing-key": {
			Type: types.ProviderOpenAI,
		},
		"unsupported": {
			Type: types.ProviderType("unknown"),
		},
	}

	form := NewAgentForm(cfg)

	model := fieldByKey(form, "model")
	require.NotNil(t, model)
	assert.Equal(t, tuicore.InputText, model.Type)
	assert.Equal(t, "manual-agent-model", model.Value)
	assert.Contains(t, model.Description, "Could not fetch models")
	assert.Contains(t, model.Description, "missing API key")

	fallbackModel := fieldByKey(form, "fallback_model")
	require.NotNil(t, fallbackModel)
	assert.Equal(t, tuicore.InputText, fallbackModel.Type)
	assert.Equal(t, "manual-fallback-model", fallbackModel.Value)
	assert.Contains(t, fallbackModel.Description, "Could not fetch models")
	assert.Contains(t, fallbackModel.Description, "unsupported type")
}

func TestMultiAgentFormWiresDefaultsVisibilityAndValidation(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.MultiAgent = false
	cfg.Agent.MaxDelegationRounds = 4
	cfg.Agent.MaxTurns = 8
	cfg.Agent.AgentsDir = "/tmp/agents"

	form := NewMultiAgentForm(cfg)

	assert.Equal(t, "Multi-Agent Configuration", form.Title)
	assert.Equal(t, []string{
		"multi_agent",
		"max_delegation_rounds",
		"max_turns",
		"error_correction_enabled",
		"agents_dir",
		"orchestration_mode",
		"orc_cb_failure_threshold",
		"orc_cb_reset_timeout",
		"orc_budget_tool_call_limit",
		"orc_budget_delegation_limit",
		"orc_budget_alert_threshold",
		"orc_recovery_max_retries",
		"orc_recovery_cooldown",
	}, formKeys(form))

	multiAgent := fieldByKey(form, "multi_agent")
	require.NotNil(t, multiAgent)
	assert.Equal(t, tuicore.InputBool, multiAgent.Type)
	assert.False(t, multiAgent.Checked)
	assert.Equal(t, []string{
		"multi_agent",
		"max_delegation_rounds",
		"max_turns",
		"error_correction_enabled",
		"agents_dir",
	}, visibleFormKeys(form))

	maxDelegationRounds := fieldByKey(form, "max_delegation_rounds")
	require.NotNil(t, maxDelegationRounds)
	assert.Equal(t, "4", maxDelegationRounds.Value)
	require.NoError(t, maxDelegationRounds.Validate("0"))
	assert.EqualError(t, maxDelegationRounds.Validate("-1"), "must be a non-negative integer")

	maxTurns := fieldByKey(form, "max_turns")
	require.NotNil(t, maxTurns)
	assert.Equal(t, "8", maxTurns.Value)
	require.NoError(t, maxTurns.Validate("12"))
	assert.EqualError(t, maxTurns.Validate("many"), "must be a non-negative integer")

	assert.True(t, fieldByKey(form, "error_correction_enabled").Checked)
	assert.Equal(t, "/tmp/agents", fieldByKey(form, "agents_dir").Value)

	mode := fieldByKey(form, "orchestration_mode")
	require.NotNil(t, mode)
	assert.Equal(t, tuicore.InputSelect, mode.Type)
	assert.Equal(t, "classic", mode.Value)
	assert.Equal(t, []string{"classic", "structured"}, mode.Options)

	multiAgent.Checked = true
	assert.Contains(t, visibleFormKeys(form), "orchestration_mode")
	assert.NotContains(t, visibleFormKeys(form), "orc_cb_failure_threshold")

	mode.Value = "structured"
	assert.Equal(t, []string{
		"multi_agent",
		"max_delegation_rounds",
		"max_turns",
		"error_correction_enabled",
		"agents_dir",
		"orchestration_mode",
		"orc_cb_failure_threshold",
		"orc_cb_reset_timeout",
		"orc_budget_tool_call_limit",
		"orc_budget_delegation_limit",
		"orc_budget_alert_threshold",
		"orc_recovery_max_retries",
		"orc_recovery_cooldown",
	}, visibleFormKeys(form))

	assert.Equal(t, "3", fieldByKey(form, "orc_cb_failure_threshold").Value)
	assert.Equal(t, "30s", fieldByKey(form, "orc_cb_reset_timeout").Value)
	assert.Equal(t, "50", fieldByKey(form, "orc_budget_tool_call_limit").Value)
	assert.Equal(t, "15", fieldByKey(form, "orc_budget_delegation_limit").Value)
	assert.Equal(t, "0.80", fieldByKey(form, "orc_budget_alert_threshold").Value)
	assert.Equal(t, "2", fieldByKey(form, "orc_recovery_max_retries").Value)
	assert.Equal(t, "5m0s", fieldByKey(form, "orc_recovery_cooldown").Value)

	require.NoError(t, fieldByKey(form, "orc_cb_failure_threshold").Validate("0"))
	assert.EqualError(
		t,
		fieldByKey(form, "orc_cb_failure_threshold").Validate("-1"),
		"must be a non-negative integer",
	)
	require.NoError(t, fieldByKey(form, "orc_cb_reset_timeout").Validate("1m"))
	assert.EqualError(
		t,
		fieldByKey(form, "orc_cb_reset_timeout").Validate("later"),
		"must be a valid duration (e.g. 30s, 1m)",
	)
	require.NoError(t, fieldByKey(form, "orc_budget_alert_threshold").Validate("1.0"))
	assert.EqualError(
		t,
		fieldByKey(form, "orc_budget_alert_threshold").Validate("1.1"),
		"must be between 0.0 and 1.0",
	)
	assert.EqualError(
		t,
		fieldByKey(form, "orc_budget_alert_threshold").Validate("high"),
		"must be a decimal number",
	)
	require.NoError(t, fieldByKey(form, "orc_recovery_cooldown").Validate("10m"))
	assert.EqualError(
		t,
		fieldByKey(form, "orc_recovery_cooldown").Validate("never"),
		"must be a valid duration (e.g. 5m, 10m)",
	)
}

func TestMultiAgentFormPreservesExplicitOrchestrationValues(t *testing.T) {
	errorCorrection := false
	cfg := config.DefaultConfig()
	cfg.Agent.MultiAgent = true
	cfg.Agent.ErrorCorrectionEnabled = &errorCorrection
	cfg.Agent.Orchestration.Mode = "structured"
	cfg.Agent.Orchestration.CircuitBreaker.FailureThreshold = 9
	cfg.Agent.Orchestration.CircuitBreaker.ResetTimeout = 17 * time.Second
	cfg.Agent.Orchestration.Budget.ToolCallLimit = 77
	cfg.Agent.Orchestration.Budget.DelegationLimit = 33
	cfg.Agent.Orchestration.Budget.AlertThreshold = 0.55
	cfg.Agent.Orchestration.Recovery.MaxRetries = 6
	cfg.Agent.Orchestration.Recovery.CircuitBreakerCooldown = 11 * time.Minute

	form := NewMultiAgentForm(cfg)

	assert.False(t, fieldByKey(form, "error_correction_enabled").Checked)
	assert.Equal(t, "structured", fieldByKey(form, "orchestration_mode").Value)
	assert.Equal(t, "9", fieldByKey(form, "orc_cb_failure_threshold").Value)
	assert.Equal(t, "17s", fieldByKey(form, "orc_cb_reset_timeout").Value)
	assert.Equal(t, "77", fieldByKey(form, "orc_budget_tool_call_limit").Value)
	assert.Equal(t, "33", fieldByKey(form, "orc_budget_delegation_limit").Value)
	assert.Equal(t, "0.55", fieldByKey(form, "orc_budget_alert_threshold").Value)
	assert.Equal(t, "6", fieldByKey(form, "orc_recovery_max_retries").Value)
	assert.Equal(t, "11m0s", fieldByKey(form, "orc_recovery_cooldown").Value)
}

func TestA2AAndPaymentFormsWireFieldsAndValidation(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.A2A.Enabled = true
	cfg.A2A.BaseURL = "https://agent.example.com"
	cfg.A2A.AgentName = "agent-one"
	cfg.A2A.AgentDescription = "Test agent"
	cfg.Payment.Enabled = true
	cfg.Payment.WalletProvider = "composite"
	cfg.Payment.Network.ChainID = 8453
	cfg.Payment.Network.RPCURL = "https://base.example"
	cfg.Payment.Network.USDCContract = "0xusdc"
	cfg.Payment.Limits.MaxPerTx = "3.00"
	cfg.Payment.Limits.MaxDaily = "12.00"
	cfg.Payment.Limits.AutoApproveBelow = "0.25"
	cfg.Payment.X402.AutoIntercept = true
	cfg.Payment.X402.MaxAutoPayAmount = "0.75"

	a2aForm := NewA2AForm(cfg)
	assert.Equal(t, "A2A Protocol Configuration", a2aForm.Title)
	assert.Equal(t, []string{
		"a2a_enabled",
		"a2a_base_url",
		"a2a_agent_name",
		"a2a_agent_desc",
	}, formKeys(a2aForm))
	assert.True(t, fieldByKey(a2aForm, "a2a_enabled").Checked)
	assert.Equal(t, "https://agent.example.com", fieldByKey(a2aForm, "a2a_base_url").Value)
	assert.Equal(t, "agent-one", fieldByKey(a2aForm, "a2a_agent_name").Value)
	assert.Equal(t, "Test agent", fieldByKey(a2aForm, "a2a_agent_desc").Value)

	paymentForm := NewPaymentForm(cfg)
	assert.Equal(t, "Payment Configuration", paymentForm.Title)
	assert.Equal(t, []string{
		"payment_enabled",
		"payment_wallet_provider",
		"payment_chain_id",
		"payment_rpc_url",
		"payment_usdc_contract",
		"payment_max_per_tx",
		"payment_max_daily",
		"payment_auto_approve",
		"payment_x402_auto",
		"payment_x402_max",
	}, formKeys(paymentForm))
	assert.True(t, fieldByKey(paymentForm, "payment_enabled").Checked)
	assert.Equal(t, "composite", fieldByKey(paymentForm, "payment_wallet_provider").Value)
	assert.Equal(t, []string{"local", "rpc", "composite"}, fieldByKey(paymentForm, "payment_wallet_provider").Options)
	assert.Equal(t, "8453", fieldByKey(paymentForm, "payment_chain_id").Value)
	assert.Equal(t, "https://base.example", fieldByKey(paymentForm, "payment_rpc_url").Value)
	assert.Equal(t, "0xusdc", fieldByKey(paymentForm, "payment_usdc_contract").Value)
	assert.Equal(t, "3.00", fieldByKey(paymentForm, "payment_max_per_tx").Value)
	assert.Equal(t, "12.00", fieldByKey(paymentForm, "payment_max_daily").Value)
	assert.Equal(t, "0.25", fieldByKey(paymentForm, "payment_auto_approve").Value)
	assert.True(t, fieldByKey(paymentForm, "payment_x402_auto").Checked)
	assert.Equal(t, "0.75", fieldByKey(paymentForm, "payment_x402_max").Value)

	chainID := fieldByKey(paymentForm, "payment_chain_id")
	require.NotNil(t, chainID)
	require.NoError(t, chainID.Validate("1"))
	assert.EqualError(t, chainID.Validate("base"), "must be an integer")
}

func visibleFormKeys(form *tuicore.FormModel) []string {
	fields := form.VisibleFields()
	keys := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == nil {
			keys = append(keys, "<nil>")
			continue
		}
		keys = append(keys, field.Key)
	}
	return keys
}
