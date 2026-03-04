package settings

import (
	"fmt"
	"strconv"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewMultiAgentForm creates the Multi-Agent configuration form.
func NewMultiAgentForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Multi-Agent Configuration")

	form.AddField(&tuicore.Field{
		Key: "multi_agent", Label: "Enable Multi-Agent Orchestration", Type: tuicore.InputBool,
		Checked:     cfg.Agent.MultiAgent,
		Description: "Allow the agent to spawn and coordinate sub-agents for complex tasks",
	})

	form.AddField(&tuicore.Field{
		Key: "max_delegation_rounds", Label: "Max Delegation Rounds", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Agent.MaxDelegationRounds),
		Placeholder: "10",
		Description: "Maximum orchestrator→sub-agent delegation rounds per turn (0 = default)",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "max_turns", Label: "Max Turns", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Agent.MaxTurns),
		Placeholder: "25",
		Description: "Maximum tool-calling iterations per agent run (0 = default)",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "error_correction_enabled", Label: "Error Correction", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.Agent.ErrorCorrectionEnabled, true),
		Description: "Enable learning-based error correction for agent tool calls",
	})

	form.AddField(&tuicore.Field{
		Key: "agents_dir", Label: "Agents Directory", Type: tuicore.InputText,
		Value:       cfg.Agent.AgentsDir,
		Placeholder: "~/.lango/agents",
		Description: "Directory containing user-defined AGENT.md files (<dir>/<name>/AGENT.md)",
	})

	return &form
}

// NewA2AForm creates the A2A Protocol configuration form.
func NewA2AForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("A2A Protocol Configuration")

	form.AddField(&tuicore.Field{
		Key: "a2a_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.A2A.Enabled,
		Description: "Enable Google A2A (Agent-to-Agent) protocol for inter-agent communication",
	})

	form.AddField(&tuicore.Field{
		Key: "a2a_base_url", Label: "Base URL", Type: tuicore.InputText,
		Value:       cfg.A2A.BaseURL,
		Placeholder: "https://your-agent.example.com",
		Description: "Public URL where this agent's A2A endpoint is accessible",
	})

	form.AddField(&tuicore.Field{
		Key: "a2a_agent_name", Label: "Agent Name", Type: tuicore.InputText,
		Value:       cfg.A2A.AgentName,
		Placeholder: "my-lango-agent",
		Description: "Human-readable name advertised in the A2A agent card",
	})

	form.AddField(&tuicore.Field{
		Key: "a2a_agent_desc", Label: "Agent Description", Type: tuicore.InputText,
		Value:       cfg.A2A.AgentDescription,
		Placeholder: "A helpful AI assistant",
		Description: "Description of this agent's capabilities for A2A discovery",
	})

	return &form
}

// NewPaymentForm creates the Payment configuration form.
func NewPaymentForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Payment Configuration")

	form.AddField(&tuicore.Field{
		Key: "payment_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Payment.Enabled,
		Description: "Enable blockchain-based USDC payment capabilities",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_wallet_provider", Label: "Wallet Provider", Type: tuicore.InputSelect,
		Value:       cfg.Payment.WalletProvider,
		Options:     []string{"local", "rpc", "composite"},
		Description: "Wallet backend: local=embedded key, rpc=remote signer, composite=multi-wallet",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_chain_id", Label: "Chain ID", Type: tuicore.InputInt,
		Value:       strconv.FormatInt(cfg.Payment.Network.ChainID, 10),
		Description: "EVM chain ID (e.g. 84532 for Base Sepolia, 8453 for Base Mainnet)",
		Validate: func(s string) error {
			if _, err := strconv.ParseInt(s, 10, 64); err != nil {
				return fmt.Errorf("must be an integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "payment_rpc_url", Label: "RPC URL", Type: tuicore.InputText,
		Value:       cfg.Payment.Network.RPCURL,
		Placeholder: "https://sepolia.base.org",
		Description: "Ethereum JSON-RPC endpoint URL for blockchain interactions",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_usdc_contract", Label: "USDC Contract", Type: tuicore.InputText,
		Value:       cfg.Payment.Network.USDCContract,
		Placeholder: "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		Description: "USDC token contract address on the selected chain",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_max_per_tx", Label: "Max Per Transaction (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.Limits.MaxPerTx,
		Placeholder: "1.00",
		Description: "Maximum USDC amount allowed per single transaction",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_max_daily", Label: "Max Daily (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.Limits.MaxDaily,
		Placeholder: "10.00",
		Description: "Maximum total USDC spending allowed per 24-hour period",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_auto_approve", Label: "Auto-Approve Below (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.Limits.AutoApproveBelow,
		Placeholder: "0.10",
		Description: "Transactions below this amount are auto-approved without user confirmation",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_x402_auto", Label: "X402 Auto-Intercept", Type: tuicore.InputBool,
		Checked:     cfg.Payment.X402.AutoIntercept,
		Description: "Automatically handle HTTP 402 Payment Required responses with USDC",
	})

	form.AddField(&tuicore.Field{
		Key: "payment_x402_max", Label: "X402 Max Auto-Pay (USDC)", Type: tuicore.InputText,
		Value:       cfg.Payment.X402.MaxAutoPayAmount,
		Placeholder: "0.50",
		Description: "Maximum USDC to auto-pay for a single X402 response",
	})

	return &form
}
