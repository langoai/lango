package payment

import (
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/x402"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuildTools_BaseSet(t *testing.T) {
	t.Parallel()

	tools := BuildTools(nil, nil, nil, 84532, nil)

	require.Len(t, tools, 5, "base set: send, balance, history, limits, wallet_info")

	names := toolNames(tools)
	for _, name := range []string{"payment_send", "payment_balance", "payment_history", "payment_limits", "payment_wallet_info"} {
		assert.Contains(t, names, name)
	}

	// Conditional tools absent without secrets/interceptor.
	assert.NotContains(t, names, "payment_create_wallet")
	assert.NotContains(t, names, "payment_x402_fetch")
}

func TestBuildTools_SafetyLevels(t *testing.T) {
	t.Parallel()

	tools := BuildTools(nil, nil, nil, 84532, nil)

	levels := make(map[string]agent.SafetyLevel, len(tools))
	for _, tool := range tools {
		levels[tool.Name] = tool.SafetyLevel
	}

	assert.Equal(t, agent.SafetyLevelDangerous, levels["payment_send"], "send must be dangerous")

	for _, name := range []string{"payment_balance", "payment_history", "payment_limits", "payment_wallet_info"} {
		assert.Equal(t, agent.SafetyLevelSafe, levels[name], "%s must be safe", name)
	}
}

func TestBuildTools_ConditionalCreateWallet(t *testing.T) {
	t.Parallel()

	// Non-nil SecretsStore adds payment_create_wallet.
	secrets := &security.SecretsStore{}
	tools := BuildTools(nil, nil, secrets, 84532, nil)

	names := toolNames(tools)
	assert.Contains(t, names, "payment_create_wallet")
	assert.NotContains(t, names, "payment_x402_fetch")
}

func TestBuildTools_ConditionalX402(t *testing.T) {
	t.Parallel()

	// Interceptor with Enabled=true adds payment_x402_fetch.
	interceptor := x402.NewInterceptor(nil, nil, x402.Config{Enabled: true}, zap.NewNop().Sugar())
	tools := BuildTools(nil, nil, nil, 84532, interceptor)

	names := toolNames(tools)
	assert.Contains(t, names, "payment_x402_fetch")
}

func TestBuildTools_AllConditional(t *testing.T) {
	t.Parallel()

	secrets := &security.SecretsStore{}
	interceptor := x402.NewInterceptor(nil, nil, x402.Config{Enabled: true}, zap.NewNop().Sugar())
	tools := BuildTools(nil, nil, secrets, 84532, interceptor)

	require.Len(t, tools, 7, "all 7 tools with both secrets and interceptor")
}

func TestBuildTools_DisabledInterceptor(t *testing.T) {
	t.Parallel()

	// Interceptor with Enabled=false does NOT add payment_x402_fetch.
	interceptor := x402.NewInterceptor(nil, nil, x402.Config{Enabled: false}, zap.NewNop().Sugar())
	tools := BuildTools(nil, nil, nil, 84532, interceptor)

	names := toolNames(tools)
	assert.NotContains(t, names, "payment_x402_fetch")
	require.Len(t, tools, 5, "disabled interceptor = base set only")
}

func TestBuildTools_HandlerParameterValidation(t *testing.T) {
	t.Parallel()

	tools := BuildTools(nil, nil, nil, 84532, nil)

	sendTool := findTool(tools, "payment_send")
	require.NotNil(t, sendTool)

	// payment_send requires to, amount, purpose — handler validates.
	_, err := sendTool.Handler(t.Context(), map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

// --- helpers ---

func toolNames(tools []*agent.Tool) map[string]bool {
	m := make(map[string]bool, len(tools))
	for _, tl := range tools {
		m[tl.Name] = true
	}
	return m
}

func findTool(tools []*agent.Tool, name string) *agent.Tool {
	for _, tl := range tools {
		if tl.Name == name {
			return tl
		}
	}
	return nil
}
