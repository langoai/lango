package settings

import (
	"fmt"
	"strconv"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewSmartAccountForm creates the Smart Account configuration form.
func NewSmartAccountForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Smart Account Configuration")

	form.AddField(&tuicore.Field{
		Key: "sa_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.SmartAccount.Enabled,
		Description: "Enable ERC-7579 modular smart account support",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_factory_address", Label: "Factory Address", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.FactoryAddress,
		Placeholder: "0x...",
		Description: "Smart account factory contract address",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_entrypoint_address", Label: "EntryPoint Address", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.EntryPointAddress,
		Placeholder: "0x...",
		Description: "ERC-4337 EntryPoint contract address",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_safe7579_address", Label: "Safe7579 Address", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.Safe7579Address,
		Placeholder: "0x...",
		Description: "Safe7579 adapter contract address",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_fallback_handler", Label: "Fallback Handler", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.FallbackHandler,
		Placeholder: "0x...",
		Description: "Fallback handler contract address",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_bundler_url", Label: "Bundler URL", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.BundlerURL,
		Placeholder: "https://bundler.example.com",
		Description: "ERC-4337 bundler RPC endpoint URL",
	})

	return &form
}

// NewSmartAccountSessionForm creates the Smart Account Session configuration form.
func NewSmartAccountSessionForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("SA Session Keys Configuration")

	form.AddField(&tuicore.Field{
		Key: "sa_session_max_duration", Label: "Max Duration", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.Session.MaxDuration.String(),
		Placeholder: "24h",
		Description: "Maximum session key validity duration",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_session_default_gas_limit", Label: "Default Gas Limit", Type: tuicore.InputText,
		Value:       strconv.FormatUint(cfg.SmartAccount.Session.DefaultGasLimit, 10),
		Placeholder: "500000",
		Description: "Default gas limit for session key transactions",
		Validate: func(s string) error {
			if _, err := strconv.ParseUint(s, 10, 64); err != nil {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "sa_session_max_active_keys", Label: "Max Active Keys", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.SmartAccount.Session.MaxActiveKeys),
		Placeholder: "10",
		Description: "Maximum number of concurrently active session keys",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewSmartAccountPaymasterForm creates the Smart Account Paymaster configuration form.
func NewSmartAccountPaymasterForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("SA Paymaster Configuration")

	form.AddField(&tuicore.Field{
		Key: "sa_paymaster_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.SmartAccount.Paymaster.Enabled,
		Description: "Enable paymaster for gasless USDC transactions",
	})

	provider := cfg.SmartAccount.Paymaster.Provider
	if provider == "" {
		provider = "circle"
	}
	form.AddField(&tuicore.Field{
		Key: "sa_paymaster_provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:       provider,
		Options:     []string{"circle", "pimlico", "alchemy"},
		Description: "Paymaster service provider",
	})

	mode := cfg.SmartAccount.Paymaster.Mode
	if mode == "" {
		mode = "rpc"
	}
	form.AddField(&tuicore.Field{
		Key: "sa_paymaster_mode", Label: "Mode", Type: tuicore.InputSelect,
		Value:       mode,
		Options:     []string{"rpc", "permit"},
		Description: "Paymaster mode: rpc (API-based) or permit (on-chain EIP-2612, no API key)",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_paymaster_rpc_url", Label: "RPC URL", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.Paymaster.RPCURL,
		Placeholder: "https://paymaster.example.com",
		Description: "Paymaster service RPC endpoint URL (required for rpc mode)",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_paymaster_token_address", Label: "Token Address", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.Paymaster.TokenAddress,
		Placeholder: "0x...",
		Description: "USDC token contract address for paymaster",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_paymaster_address", Label: "Paymaster Address", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.Paymaster.PaymasterAddress,
		Placeholder: "0x...",
		Description: "Paymaster contract address",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_paymaster_policy_id", Label: "Policy ID", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.Paymaster.PolicyID,
		Placeholder: "policy-id-from-provider",
		Description: "Paymaster policy identifier (provider-specific)",
	})

	fallbackMode := cfg.SmartAccount.Paymaster.FallbackMode
	if fallbackMode == "" {
		fallbackMode = "abort"
	}
	form.AddField(&tuicore.Field{
		Key: "sa_paymaster_fallback_mode", Label: "Fallback Mode", Type: tuicore.InputSelect,
		Value:       fallbackMode,
		Options:     []string{"abort", "direct"},
		Description: "Behavior when paymaster is unavailable: abort or pay directly",
	})

	return &form
}

// NewSmartAccountModulesForm creates the Smart Account Modules configuration form.
func NewSmartAccountModulesForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("SA Modules Configuration")

	form.AddField(&tuicore.Field{
		Key: "sa_modules_session_validator", Label: "Session Validator", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.Modules.SessionValidatorAddress,
		Placeholder: "0x...",
		Description: "Session key validator module contract address",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_modules_spending_hook", Label: "Spending Hook", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.Modules.SpendingHookAddress,
		Placeholder: "0x...",
		Description: "Spending limit hook module contract address",
	})

	form.AddField(&tuicore.Field{
		Key: "sa_modules_escrow_executor", Label: "Escrow Executor", Type: tuicore.InputText,
		Value:       cfg.SmartAccount.Modules.EscrowExecutorAddress,
		Placeholder: "0x...",
		Description: "Escrow executor module contract address",
	})

	return &form
}
