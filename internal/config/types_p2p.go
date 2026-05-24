package config

import "time"

// P2PConfig defines peer-to-peer network settings for the Sovereign Agent Network.
type P2PConfig struct {
	// Enable P2P networking.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// ListenAddrs are the multiaddrs to listen on (e.g. /ip4/0.0.0.0/tcp/9000).
	ListenAddrs []string `mapstructure:"listenAddrs" json:"listenAddrs"`

	// BootstrapPeers are initial peers to connect to for DHT bootstrapping.
	BootstrapPeers []string `mapstructure:"bootstrapPeers" json:"bootstrapPeers"`

	// Deprecated: KeyDir is the legacy directory for persisting node keys.
	// Node keys are now stored in SecretsStore (encrypted) when available.
	// This field is retained for backward compatibility and migration.
	KeyDir string `mapstructure:"keyDir" json:"keyDir,omitempty"`

	// EnableRelay allows this node to act as a relay for NAT traversal.
	EnableRelay bool `mapstructure:"enableRelay" json:"enableRelay"`

	// EnableMDNS enables multicast DNS for local peer discovery.
	EnableMDNS bool `mapstructure:"enableMdns" json:"enableMdns"`

	// MaxPeers is the maximum number of connected peers.
	MaxPeers int `mapstructure:"maxPeers" json:"maxPeers"`

	// HandshakeTimeout is the maximum duration for peer handshake.
	HandshakeTimeout time.Duration `mapstructure:"handshakeTimeout" json:"handshakeTimeout"`

	// SessionTokenTTL is the lifetime of session tokens after handshake.
	SessionTokenTTL time.Duration `mapstructure:"sessionTokenTtl" json:"sessionTokenTtl"`

	// AutoApproveKnownPeers skips HITL approval for previously authenticated peers.
	AutoApproveKnownPeers bool `mapstructure:"autoApproveKnownPeers" json:"autoApproveKnownPeers"`

	// FirewallRules defines static ACL rules for the knowledge firewall.
	FirewallRules []FirewallRule `mapstructure:"firewallRules" json:"firewallRules"`

	// GossipInterval is the interval for gossip-based agent card propagation.
	GossipInterval time.Duration `mapstructure:"gossipInterval" json:"gossipInterval"`

	// ZKHandshake enables ZK-enhanced handshake instead of plain signature mode.
	ZKHandshake bool `mapstructure:"zkHandshake" json:"zkHandshake"`

	// ZKAttestation enables ZK attestation proofs on responses to peers.
	ZKAttestation bool `mapstructure:"zkAttestation" json:"zkAttestation"`

	// ZKP holds zero-knowledge proof settings.
	ZKP ZKPConfig `mapstructure:"zkp" json:"zkp"`

	// Pricing for paid P2P tool invocations.
	Pricing P2PPricingConfig `mapstructure:"pricing" json:"pricing"`

	// OwnerProtection prevents owner PII from leaking via P2P.
	OwnerProtection OwnerProtectionConfig `mapstructure:"ownerProtection" json:"ownerProtection"`

	// MinTrustScore is the minimum reputation to accept requests (0.0 to 1.0, default 0.3).
	// This is the admission gate and is separate from post-pay payment trust thresholds.
	MinTrustScore float64 `mapstructure:"minTrustScore" json:"minTrustScore"`

	// ToolIsolation configures process isolation for remote tool invocations.
	ToolIsolation ToolIsolationConfig `mapstructure:"toolIsolation" json:"toolIsolation"`

	// RequireSignedChallenge rejects unsigned challenges from peers when true.
	// When false (default), unsigned legacy challenges are accepted for backward compatibility.
	RequireSignedChallenge bool `mapstructure:"requireSignedChallenge" json:"requireSignedChallenge"`

	// EnablePQHandshake enables post-quantum hybrid KEM (X25519-MLKEM768)
	// key exchange during peer handshake. When true, protocol v1.2 is
	// advertised and PQ session keys are derived. Default: false (opt-in).
	EnablePQHandshake bool `mapstructure:"enablePqHandshake" json:"enablePqHandshake"`

	// Workspace configures collaborative workspace settings for P2P agent co-work.
	Workspace WorkspaceConfig `mapstructure:"workspace" json:"workspace"`

	// Team configures team health monitoring and membership policies.
	Team TeamConfig `mapstructure:"team" json:"team"`

	// MaxSafetyLevel is the highest SafetyLevel a P2P peer may invoke.
	// Tools above this level are rejected. Valid values: "safe", "moderate", "dangerous".
	// Default: "moderate" — blocks Dangerous tools from remote peers.
	MaxSafetyLevel string `mapstructure:"maxSafetyLevel" json:"maxSafetyLevel"`

	// AllowedTools is an explicit whitelist of tool names that bypass the
	// SafetyLevel gate for P2P peers. An empty list means the SafetyLevel
	// gate alone decides (together with firewall ACL).
	AllowedTools []string `mapstructure:"allowedTools" json:"allowedTools,omitempty"`
}

// GetKeyDir returns the legacy directory for persisting node keys.
func (c P2PConfig) GetKeyDir() string { return c.KeyDir }

// GetMaxPeers returns the maximum number of connected peers.
func (c P2PConfig) GetMaxPeers() int { return c.MaxPeers }

// GetListenAddrs returns the multiaddrs to listen on.
func (c P2PConfig) GetListenAddrs() []string { return c.ListenAddrs }

// GetEnableRelay reports whether this node acts as a relay for NAT traversal.
func (c P2PConfig) GetEnableRelay() bool { return c.EnableRelay }

// GetBootstrapPeers returns the initial peers for DHT bootstrapping.
func (c P2PConfig) GetBootstrapPeers() []string { return c.BootstrapPeers }

// GetEnableMDNS reports whether multicast DNS discovery is enabled.
func (c P2PConfig) GetEnableMDNS() bool { return c.EnableMDNS }

// ToolIsolationConfig configures subprocess isolation for P2P tool execution.
type ToolIsolationConfig struct {
	// Enabled turns on subprocess isolation for remote peer tool invocations.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// TimeoutPerTool is the maximum duration for a single tool execution (default: 30s).
	TimeoutPerTool time.Duration `mapstructure:"timeoutPerTool" json:"timeoutPerTool"`

	// MaxMemoryMB is a soft memory limit per subprocess in megabytes (Phase 2).
	MaxMemoryMB int `mapstructure:"maxMemoryMB" json:"maxMemoryMB"`

	// Container configures container-based tool execution sandbox (Phase 2).
	Container ContainerSandboxConfig `mapstructure:"container" json:"container"`
}

// ContainerSandboxConfig configures container-based tool execution isolation.
type ContainerSandboxConfig struct {
	// Enabled activates container-based sandbox instead of subprocess isolation.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// RequireContainer enforces fail-closed behavior: when true and no container
	// runtime (Docker/gVisor) is available, tool execution is refused instead of
	// silently falling back to NativeRuntime. Defaults to true for security.
	RequireContainer bool `mapstructure:"requireContainer" json:"requireContainer"`

	// Runtime selects the container runtime: "auto", "docker", "gvisor", or "native" (default: "auto").
	Runtime string `mapstructure:"runtime" json:"runtime"`

	// Image is the Docker image for the sandbox container (default: "lango-sandbox:latest").
	Image string `mapstructure:"image" json:"image"`

	// NetworkMode is the Docker network mode for sandbox containers (default: "none").
	NetworkMode string `mapstructure:"networkMode" json:"networkMode"`

	// ReadOnlyRootfs mounts the container root filesystem as read-only (default: true).
	ReadOnlyRootfs *bool `mapstructure:"readOnlyRootfs" json:"readOnlyRootfs"`

	// CPUQuotaUS is the Docker CPU quota in microseconds (0 = unlimited).
	CPUQuotaUS int64 `mapstructure:"cpuQuotaUs" json:"cpuQuotaUs"`

	// PoolSize is the number of pre-warmed containers in the pool (0 = disabled).
	PoolSize int `mapstructure:"poolSize" json:"poolSize"`

	// PoolIdleTimeout is the idle timeout before pool containers are recycled (default: 5m).
	PoolIdleTimeout time.Duration `mapstructure:"poolIdleTimeout" json:"poolIdleTimeout"`
}

// P2PPricingConfig defines pricing for P2P tool invocations.
type P2PPricingConfig struct {
	// Enable paid tool invocations.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// PerQuery is the default price per query in USDC (e.g. "0.50").
	PerQuery string `mapstructure:"perQuery" json:"perQuery"`

	// ToolPrices maps tool names to their specific prices in USDC.
	ToolPrices map[string]string `mapstructure:"toolPrices" json:"toolPrices,omitempty"`

	// TrustThresholds configures trust-based payment tier thresholds.
	// These thresholds are separate from MinTrustScore admission checks.
	TrustThresholds TrustThresholds `mapstructure:"trustThresholds" json:"trustThresholds"`

	// Settlement configures on-chain settlement behavior.
	Settlement SettlementConfig `mapstructure:"settlement" json:"settlement"`
}

// TrustThresholds defines score thresholds for payment tier routing.
type TrustThresholds struct {
	// PostPayMinScore is the minimum reputation score for post-pay eligibility (default: 0.8).
	PostPayMinScore float64 `mapstructure:"postPayMinScore" json:"postPayMinScore"`
}

// SettlementConfig configures on-chain settlement parameters.
type SettlementConfig struct {
	// ReceiptTimeout is the maximum wait time for on-chain receipt confirmation (default: 2m).
	ReceiptTimeout time.Duration `mapstructure:"receiptTimeout" json:"receiptTimeout"`

	// MaxRetries is the maximum number of submission retries (default: 3).
	MaxRetries int `mapstructure:"maxRetries" json:"maxRetries"`
}

// OwnerProtectionConfig configures owner data protection for P2P responses.
type OwnerProtectionConfig struct {
	// OwnerName is the owner's name to block from P2P responses.
	OwnerName string `mapstructure:"ownerName" json:"ownerName"`

	// OwnerEmail is the owner's email to block from P2P responses.
	OwnerEmail string `mapstructure:"ownerEmail" json:"ownerEmail"`

	// OwnerPhone is the owner's phone number to block from P2P responses.
	OwnerPhone string `mapstructure:"ownerPhone" json:"ownerPhone"`

	// ExtraTerms are additional terms to block from P2P responses.
	ExtraTerms []string `mapstructure:"extraTerms" json:"extraTerms,omitempty"`

	// BlockConversations blocks all conversation-related fields from P2P responses (default: true).
	BlockConversations *bool `mapstructure:"blockConversations" json:"blockConversations"`
}

// ZKPConfig defines zero-knowledge proof settings.
type ZKPConfig struct {
	// ProofCacheDir is the directory for caching compiled circuits and proving keys.
	ProofCacheDir string `mapstructure:"proofCacheDir" json:"proofCacheDir"`

	// ProvingScheme selects the ZKP proving scheme: "plonk" or "groth16".
	ProvingScheme string `mapstructure:"provingScheme" json:"provingScheme"`

	// SRSMode selects the SRS generation mode: "unsafe" (default) or "file".
	SRSMode string `mapstructure:"srsMode" json:"srsMode"`

	// SRSPath is the path to the SRS file (used when SRSMode == "file").
	SRSPath string `mapstructure:"srsPath" json:"srsPath"`

	// MaxCredentialAge is the maximum age for ZK credentials (e.g. "24h").
	MaxCredentialAge string `mapstructure:"maxCredentialAge" json:"maxCredentialAge"`
}

// WorkspaceConfig configures collaborative workspace settings for P2P agent co-work.
type WorkspaceConfig struct {
	// Enabled turns on collaborative workspaces.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// DataDir is the directory for storing workspace data.
	DataDir string `mapstructure:"dataDir" json:"dataDir"`

	// MaxWorkspaces is the maximum number of concurrent workspaces.
	MaxWorkspaces int `mapstructure:"maxWorkspaces" json:"maxWorkspaces"`

	// MaxBundleSizeBytes is the maximum size of a workspace bundle in bytes.
	MaxBundleSizeBytes int64 `mapstructure:"maxBundleSizeBytes" json:"maxBundleSizeBytes"`

	// ChroniclerEnabled turns on workspace activity chronicling.
	ChroniclerEnabled bool `mapstructure:"chroniclerEnabled" json:"chroniclerEnabled"`

	// AutoSandbox automatically sandboxes workspace operations.
	AutoSandbox bool `mapstructure:"autoSandbox" json:"autoSandbox"`

	// ContributionTracking enables tracking of per-agent contributions.
	ContributionTracking bool `mapstructure:"contributionTracking" json:"contributionTracking"`

	// EnableIncrementalBundle enables incremental bundle creation (base..HEAD) instead of full repo bundles.
	EnableIncrementalBundle bool `mapstructure:"enableIncrementalBundle" json:"enableIncrementalBundle"`

	// BranchPerTask creates isolated task/{id} branches for each agent task to avoid conflicts.
	BranchPerTask bool `mapstructure:"branchPerTask" json:"branchPerTask"`

	// MaxIncrementalBundleSizeBytes is the maximum size of an incremental bundle in bytes.
	MaxIncrementalBundleSizeBytes int64 `mapstructure:"maxIncrementalBundleSizeBytes" json:"maxIncrementalBundleSizeBytes"`
}

// TeamConfig configures team health monitoring and membership policies.
type TeamConfig struct {
	// HealthCheckInterval is the interval between team health pings (default: 30s).
	HealthCheckInterval time.Duration `mapstructure:"healthCheckInterval" json:"healthCheckInterval"`

	// MaxMissedHeartbeats is the maximum consecutive missed pings before a member is unhealthy (default: 3).
	MaxMissedHeartbeats int `mapstructure:"maxMissedHeartbeats" json:"maxMissedHeartbeats"`

	// MinReputationScore is the minimum reputation to remain on a team (default: 0.3).
	MinReputationScore float64 `mapstructure:"minReputationScore" json:"minReputationScore"`

	// GitStateTracking enables tracking git HEAD hashes from health ping responses.
	GitStateTracking bool `mapstructure:"gitStateTracking" json:"gitStateTracking"`

	// AutoSyncOnDivergence automatically triggers sync when git state divergence is detected.
	AutoSyncOnDivergence bool `mapstructure:"autoSyncOnDivergence" json:"autoSyncOnDivergence"`
}

// FirewallRule defines an ACL rule for the knowledge firewall.
type FirewallRule struct {
	// PeerDID is the DID of the peer this rule applies to ("*" for all).
	PeerDID string `mapstructure:"peerDid" json:"peerDid"`

	// Action is "allow" or "deny".
	Action string `mapstructure:"action" json:"action"`

	// Tools lists tool name patterns this rule applies to.
	Tools []string `mapstructure:"tools" json:"tools"`

	// RateLimit is the maximum requests per minute (0 = unlimited).
	RateLimit int `mapstructure:"rateLimit" json:"rateLimit"`
}
