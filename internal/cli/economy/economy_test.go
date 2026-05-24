package economy

import (
	"bytes"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func executeEconomyCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestEscrowSentinelStatusCmd_UsesExistingAlertSurfaceGuidance(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = true
	cfg.Economy.Escrow.OnChain.Enabled = true

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "sentinel", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "The sentinel engine runs within the application server.")
	assert.Contains(t, out, "Use 'lango serve' to start the application server, then inspect detected alerts via the sentinel_alerts agent tool.")
	assert.NotContains(t, out, "lango economy escrow sentinel alerts")
}

func TestEscrowShowCmd_WithID_UsesEscrowStatusGuidance(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = true

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "show", "--id", "escrow-123")
	require.NoError(t, err)

	assert.Contains(t, out, "Escrow ID \"escrow-123\": use 'lango serve' and the escrow_status agent tool for live data")
	assert.NotContains(t, out, "escrow agent tools")
}

func TestEscrowShowCmd_HelpUsesTruthfulIDFlagDescription(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeEconomyCmd(t, cmd, "escrow", "show", "--help")
	assert.NoError(t, err)
	assert.Contains(t, out, "Optional escrow ID; prints live-status runtime guidance for that escrow")
	assert.NotContains(t, out, "future use")
}

func TestNewEconomyCmd_HelpExamplesMatchCurrentSurface(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })

	out, err := executeEconomyCmd(t, cmd, "--help")
	require.NoError(t, err)

	assert.Contains(t, out, "lango economy budget status --task-id=task-1")
	assert.Contains(t, out, "lango economy risk status")
	assert.Contains(t, out, "lango economy pricing status")
	assert.Contains(t, out, "lango economy negotiate status")
	assert.Contains(t, out, "lango economy escrow show")
	assert.NotContains(t, out, "lango economy risk assess")
	assert.NotContains(t, out, "lango economy pricing quote")
	assert.NotContains(t, out, "lango economy negotiate list")
	assert.NotContains(t, out, "lango economy escrow status --escrow-id")
}

func TestPricingStatusCmd_WritesEnabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Pricing.Enabled = true
	cfg.Economy.Pricing.TrustDiscount = 0.2
	cfg.Economy.Pricing.VolumeDiscount = 0.1
	cfg.Economy.Pricing.MinPrice = "0.01"

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "pricing", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Pricing Configuration:")
	assert.Contains(t, out, "Trust Discount:  20%")
	assert.Contains(t, out, "Volume Discount: 10%")
	assert.Contains(t, out, "Min Price:       0.01 USDC")
}

func TestPricingStatusCmd_WritesDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Pricing.Enabled = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "pricing", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Dynamic pricing is disabled.")
}

func TestRiskStatusCmd_WritesEnabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Risk.EscrowThreshold = "5.00"
	cfg.Economy.Risk.HighTrustScore = 0.8
	cfg.Economy.Risk.MediumTrustScore = 0.5

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "risk", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Risk Configuration:")
	assert.Contains(t, out, "Escrow Threshold: 5.00 USDC")
	assert.Contains(t, out, "High Trust Score: 0.80")
	assert.Contains(t, out, "Med Trust Score:  0.50")
}

func TestRiskStatusCmd_WritesDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "risk", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Economy layer is disabled.")
}

func TestNegotiateStatusCmd_WritesEnabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Negotiate.Enabled = true
	cfg.Economy.Negotiate.MaxRounds = 5
	cfg.Economy.Negotiate.Timeout = 30 * time.Second
	cfg.Economy.Negotiate.AutoNegotiate = true
	cfg.Economy.Negotiate.MaxDiscount = 0.3

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "negotiate", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Negotiation Configuration:")
	assert.Contains(t, out, "Max Rounds:     5")
	assert.Contains(t, out, "Timeout:        30s")
	assert.Contains(t, out, "Auto Negotiate: true")
	assert.Contains(t, out, "Max Discount:   30%")
}

func TestNegotiateStatusCmd_WritesDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Negotiate.Enabled = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "negotiate", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Negotiation is disabled.")
}

func TestBudgetStatusCmd_WritesEnabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Budget.DefaultMax = "10.00"
	cfg.Economy.Budget.AlertThresholds = []float64{0.5, 0.8, 0.95}
	hardLimit := true
	cfg.Economy.Budget.HardLimit = &hardLimit

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "budget", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Budget Configuration:")
	assert.Contains(t, out, "Default Max:      10.00 USDC")
	assert.Contains(t, out, "Alert Thresholds: [0.5 0.8 0.95]")
	assert.Contains(t, out, "Hard Limit:       enabled")
}

func TestBudgetStatusCmd_WritesTaskGuidanceToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "budget", "status", "--task-id=task-1")
	require.NoError(t, err)

	assert.Contains(t, out, `Task "task-1" budget: use 'lango serve' and economy_budget_status tool for live data`)
}

func TestBudgetStatusCmd_WritesDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "budget", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Economy layer is disabled. Enable with economy.enabled=true")
}

func TestEscrowStatusCmd_WritesEnabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = true
	cfg.Economy.Escrow.DefaultTimeout = 24 * time.Hour
	cfg.Economy.Escrow.MaxMilestones = 10
	cfg.Economy.Escrow.AutoRelease = true
	cfg.Economy.Escrow.DisputeWindow = 48 * time.Hour

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Escrow Configuration:")
	assert.Contains(t, out, "Default Timeout: 24h0m0s")
	assert.Contains(t, out, "Max Milestones:  10")
	assert.Contains(t, out, "Auto Release:    true")
	assert.Contains(t, out, "Dispute Window:  48h0m0s")
}

func TestEscrowStatusCmd_WritesDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Escrow is disabled.")
}

func TestEscrowListCmd_WritesSummaryToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = true
	cfg.Economy.Escrow.OnChain.Enabled = true
	cfg.Economy.Escrow.OnChain.Mode = "hub"
	cfg.Economy.Escrow.OnChain.HubAddress = "0x1234"
	cfg.Economy.Escrow.DefaultTimeout = 24 * time.Hour
	cfg.Economy.Escrow.AutoRelease = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "list")
	require.NoError(t, err)

	assert.Contains(t, out, "Escrow Summary:")
	assert.Contains(t, out, "On-Chain Escrow:  enabled")
	assert.Contains(t, out, "Mode:             hub")
	assert.Contains(t, out, "Hub Address:      0x1234")
	assert.Contains(t, out, "Use 'lango economy escrow show' for detailed on-chain configuration.")
}

func TestEscrowListCmd_WritesEconomyDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "list")
	require.NoError(t, err)

	assert.Contains(t, out, "Economy layer is disabled. Enable with economy.enabled=true")
}

func TestEscrowListCmd_WritesEscrowDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "list")
	require.NoError(t, err)

	assert.Contains(t, out, "Escrow is disabled. Enable with economy.escrow.enabled=true")
}

func TestEscrowShowCmd_WritesDetailedConfigToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = true
	cfg.Economy.Escrow.OnChain.Enabled = true
	cfg.Economy.Escrow.OnChain.Mode = "hub"
	cfg.Economy.Escrow.OnChain.HubAddress = "0x1234"
	cfg.Economy.Escrow.OnChain.ArbitratorAddress = "0x5678"
	cfg.Economy.Escrow.OnChain.TokenAddress = "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
	cfg.Economy.Escrow.OnChain.PollInterval = 15 * time.Second
	cfg.Economy.Escrow.Settlement.ReceiptTimeout = 2 * time.Minute
	cfg.Economy.Escrow.Settlement.MaxRetries = 3

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "show")
	require.NoError(t, err)

	assert.Contains(t, out, "On-Chain Escrow Configuration:")
	assert.Contains(t, out, "Hub Address:          0x1234")
	assert.Contains(t, out, "Arbitrator:           0x5678")
	assert.Contains(t, out, "Settlement:")
	assert.Contains(t, out, "Receipt Timeout:      2m0s")
}

func TestEscrowShowCmd_WritesDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "show")
	require.NoError(t, err)

	assert.Contains(t, out, "Escrow is disabled.")
}

func TestEscrowSentinelStatusCmd_WritesOnChainDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = true
	cfg.Economy.Escrow.OnChain.Enabled = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "sentinel", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "On-chain escrow is disabled. Sentinel monitors on-chain events.")
}

func TestEscrowSentinelStatusCmd_WritesEscrowDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Economy.Enabled = true
	cfg.Economy.Escrow.Enabled = false

	cmd := NewEconomyCmd(func() (*config.Config, error) { return cfg, nil })
	out, err := executeEconomyCmd(t, cmd, "escrow", "sentinel", "status")
	require.NoError(t, err)

	assert.Contains(t, out, "Escrow is disabled. Sentinel is not active.")
}
