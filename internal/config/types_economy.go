package config

import "time"

// EconomyConfig defines P2P economy layer settings (budget, risk, escrow, pricing, negotiation).
type EconomyConfig struct {
	// Enabled activates the economy layer.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Budget controls per-task spending limits.
	Budget BudgetConfig `mapstructure:"budget" json:"budget"`

	// Risk configures trust-based payment strategy routing.
	Risk RiskConfig `mapstructure:"risk" json:"risk"`

	// Negotiate configures the P2P negotiation protocol.
	Negotiate NegotiationConfig `mapstructure:"negotiate" json:"negotiate"`

	// Escrow configures the milestone-based escrow service.
	Escrow EscrowConfig `mapstructure:"escrow" json:"escrow"`

	// Pricing configures dynamic pricing adjustments.
	Pricing DynamicPricingConfig `mapstructure:"pricing" json:"pricing"`
}

// BudgetConfig defines per-task spending limits.
type BudgetConfig struct {
	// DefaultMax is the default maximum budget per task in USDC (e.g. "10.00").
	DefaultMax string `mapstructure:"defaultMax" json:"defaultMax"`

	// AlertThresholds are percentage thresholds that trigger budget alerts (e.g. [0.5, 0.8, 0.95]).
	AlertThresholds []float64 `mapstructure:"alertThresholds" json:"alertThresholds"`

	// HardLimit enforces budget as a hard cap (rejects overspend). Default: true.
	HardLimit *bool `mapstructure:"hardLimit" json:"hardLimit"`
}

// RiskConfig defines trust-based payment strategy routing thresholds.
type RiskConfig struct {
	// EscrowThreshold is the USDC amount above which escrow is forced (e.g. "5.00").
	EscrowThreshold string `mapstructure:"escrowThreshold" json:"escrowThreshold"`

	// HighTrustScore is the minimum trust score for DirectPay strategy (default: 0.8).
	HighTrustScore float64 `mapstructure:"highTrustScore" json:"highTrustScore"`

	// MediumTrustScore is the minimum trust score for non-ZK strategies (default: 0.5).
	MediumTrustScore float64 `mapstructure:"mediumTrustScore" json:"mediumTrustScore"`
}

// NegotiationConfig defines P2P price negotiation settings.
type NegotiationConfig struct {
	// Enabled activates the P2P negotiation protocol.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// MaxRounds is the maximum number of counter-offers (default: 5).
	MaxRounds int `mapstructure:"maxRounds" json:"maxRounds"`

	// Timeout is the negotiation session timeout (default: 5m).
	Timeout time.Duration `mapstructure:"timeout" json:"timeout"`

	// AutoNegotiate enables automatic counter-offer generation.
	AutoNegotiate bool `mapstructure:"autoNegotiate" json:"autoNegotiate"`

	// MaxDiscount is the maximum discount percentage for auto-negotiation (0-1, default: 0.2).
	MaxDiscount float64 `mapstructure:"maxDiscount" json:"maxDiscount"`
}

// EscrowConfig defines milestone-based escrow settings.
type EscrowConfig struct {
	// Enabled activates the escrow service.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// DefaultTimeout is the escrow expiration timeout (default: 24h).
	DefaultTimeout time.Duration `mapstructure:"defaultTimeout" json:"defaultTimeout"`

	// MaxMilestones is the maximum milestones per escrow (default: 10).
	MaxMilestones int `mapstructure:"maxMilestones" json:"maxMilestones"`

	// AutoRelease releases funds automatically when all milestones are met.
	AutoRelease bool `mapstructure:"autoRelease" json:"autoRelease"`

	// DisputeWindow is the time window for raising disputes after completion (default: 1h).
	DisputeWindow time.Duration `mapstructure:"disputeWindow" json:"disputeWindow"`

	// Settlement configures on-chain settlement for escrow.
	Settlement EscrowSettlementConfig `mapstructure:"settlement" json:"settlement"`

	// OnChain configures the on-chain escrow hub/vault system.
	OnChain EscrowOnChainConfig `mapstructure:"onChain" json:"onChain"`
}

// EscrowOnChainConfig configures on-chain escrow contract integration.
type EscrowOnChainConfig struct {
	// Enabled activates on-chain escrow mode.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Mode selects the on-chain escrow pattern: "hub" or "vault".
	Mode string `mapstructure:"mode" json:"mode"`

	// ContractVersion selects the contract version: "v1" or "v2" (default: auto-detect).
	ContractVersion string `mapstructure:"contractVersion" json:"contractVersion"`

	// HubAddress is the deployed LangoEscrowHub contract address.
	HubAddress string `mapstructure:"hubAddress" json:"hubAddress"`

	// HubV2Address is the deployed LangoEscrowHubV2 proxy address (UUPS).
	HubV2Address string `mapstructure:"hubV2Address" json:"hubV2Address"`

	// VaultFactoryAddress is the deployed LangoVaultFactory contract address.
	VaultFactoryAddress string `mapstructure:"vaultFactoryAddress" json:"vaultFactoryAddress"`

	// VaultImplementation is the LangoVault implementation address for cloning.
	VaultImplementation string `mapstructure:"vaultImplementation" json:"vaultImplementation"`

	// BeaconAddress is the UpgradeableBeacon address for V2 vaults.
	BeaconAddress string `mapstructure:"beaconAddress" json:"beaconAddress"`

	// BeaconFactoryAddress is the LangoBeaconVaultFactory address for V2 vaults.
	BeaconFactoryAddress string `mapstructure:"beaconFactoryAddress" json:"beaconFactoryAddress"`

	// DirectSettlerAddress is the deployed DirectSettler contract address (V2).
	DirectSettlerAddress string `mapstructure:"directSettlerAddress" json:"directSettlerAddress"`

	// MilestoneSettlerAddress is the deployed MilestoneSettler contract address (V2).
	MilestoneSettlerAddress string `mapstructure:"milestoneSettlerAddress" json:"milestoneSettlerAddress"`

	// ArbitratorAddress is the dispute arbitrator address.
	ArbitratorAddress string `mapstructure:"arbitratorAddress" json:"arbitratorAddress"`

	// PollInterval is the event monitor polling interval (default: 15s).
	PollInterval time.Duration `mapstructure:"pollInterval" json:"pollInterval"`

	// ConfirmationDepth is the number of blocks to wait before processing events
	// to protect against L2 reorgs (default: 2 for Base L2).
	ConfirmationDepth uint64 `mapstructure:"confirmationDepth" json:"confirmationDepth"`

	// TokenAddress is the ERC-20 token (USDC) contract address.
	TokenAddress string `mapstructure:"tokenAddress" json:"tokenAddress"`
}

// IsV2 returns true if the on-chain config uses V2 contracts.
// Auto-detects based on HubV2Address presence when ContractVersion is empty.
func (c EscrowOnChainConfig) IsV2() bool {
	if c.ContractVersion == "v2" {
		return true
	}
	if c.ContractVersion == "v1" {
		return false
	}
	return c.HubV2Address != "" || c.BeaconFactoryAddress != ""
}

// EscrowSettlementConfig configures on-chain settlement parameters for escrow.
type EscrowSettlementConfig struct {
	// ReceiptTimeout is the maximum wait for on-chain confirmation (default: 2m).
	ReceiptTimeout time.Duration `mapstructure:"receiptTimeout" json:"receiptTimeout"`

	// MaxRetries is the maximum transaction submission attempts (default: 3).
	MaxRetries int `mapstructure:"maxRetries" json:"maxRetries"`
}

// DynamicPricingConfig defines dynamic pricing adjustment settings.
type DynamicPricingConfig struct {
	// Enabled activates dynamic pricing.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// TrustDiscount is the max discount for high-trust peers (0-1, default: 0.1).
	TrustDiscount float64 `mapstructure:"trustDiscount" json:"trustDiscount"`

	// VolumeDiscount is the max discount for high-volume peers (0-1, default: 0.05).
	VolumeDiscount float64 `mapstructure:"volumeDiscount" json:"volumeDiscount"`

	// MinPrice is the minimum price floor in USDC (e.g. "0.01").
	MinPrice string `mapstructure:"minPrice" json:"minPrice"`
}
