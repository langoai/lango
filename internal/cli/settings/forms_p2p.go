package settings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewP2PForm creates the P2P Network configuration form.
func NewP2PForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Network Configuration")

	form.AddField(&tuicore.Field{
		Key: "p2p_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.P2P.Enabled,
		Description: "Enable libp2p-based peer-to-peer networking for agent discovery",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_listen_addrs", Label: "Listen Addresses", Type: tuicore.InputText,
		Value:       strings.Join(cfg.P2P.ListenAddrs, ","),
		Placeholder: "/ip4/0.0.0.0/tcp/9000 (comma-separated)",
		Description: "Multiaddr listen addresses for incoming P2P connections",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_bootstrap_peers", Label: "Bootstrap Peers", Type: tuicore.InputText,
		Value:       strings.Join(cfg.P2P.BootstrapPeers, ","),
		Placeholder: "/ip4/host/tcp/port/p2p/peerID (comma-separated)",
		Description: "Initial peers to connect to for network discovery",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_enable_relay", Label: "Enable Relay", Type: tuicore.InputBool,
		Checked:     cfg.P2P.EnableRelay,
		Description: "Allow relaying connections for peers behind NAT",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_enable_mdns", Label: "Enable mDNS", Type: tuicore.InputBool,
		Checked:     cfg.P2P.EnableMDNS,
		Description: "Use multicast DNS for local network peer discovery",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_max_peers", Label: "Max Peers", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.P2P.MaxPeers),
		Description: "Maximum number of simultaneous peer connections",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_handshake_timeout", Label: "Handshake Timeout", Type: tuicore.InputText,
		Value:       cfg.P2P.HandshakeTimeout.String(),
		Placeholder: "30s",
		Description: "Maximum time to wait for peer handshake completion",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_session_token_ttl", Label: "Session Token TTL", Type: tuicore.InputText,
		Value:       cfg.P2P.SessionTokenTTL.String(),
		Placeholder: "24h",
		Description: "Lifetime of P2P session tokens before re-authentication is required",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_auto_approve", Label: "Auto-Approve Known Peers", Type: tuicore.InputBool,
		Checked:     cfg.P2P.AutoApproveKnownPeers,
		Description: "Skip approval for previously authenticated and trusted peers",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_gossip_interval", Label: "Gossip Interval", Type: tuicore.InputText,
		Value:       cfg.P2P.GossipInterval.String(),
		Placeholder: "30s",
		Description: "Interval between gossip protocol broadcasts for peer discovery",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_zk_handshake", Label: "ZK Handshake", Type: tuicore.InputBool,
		Checked:     cfg.P2P.ZKHandshake,
		Description: "Use zero-knowledge proofs during peer handshake for privacy",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_zk_attestation", Label: "ZK Attestation", Type: tuicore.InputBool,
		Checked:     cfg.P2P.ZKAttestation,
		Description: "Require ZK attestation proofs for tool execution results",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_require_signed_challenge", Label: "Require Signed Challenge", Type: tuicore.InputBool,
		Checked:     cfg.P2P.RequireSignedChallenge,
		Description: "Require cryptographic challenge-response during peer authentication",
	})

	form.AddField(&tuicore.Field{
		Key: "p2p_min_trust_score", Label: "Min Trust Score", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.1f", cfg.P2P.MinTrustScore),
		Placeholder: "0.3 (0.0 to 1.0)",
		Description: "Minimum trust score (0.0-1.0) required to interact with a peer",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 1.0 {
				return fmt.Errorf("must be between 0.0 and 1.0")
			}
			return nil
		},
	})

	return &form
}

// NewP2PZKPForm creates the P2P ZKP configuration form.
func NewP2PZKPForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P ZKP Configuration")

	form.AddField(&tuicore.Field{
		Key: "zkp_proof_cache_dir", Label: "Proof Cache Directory", Type: tuicore.InputText,
		Value:       cfg.P2P.ZKP.ProofCacheDir,
		Placeholder: "~/.lango/p2p/zkp-cache",
		Description: "Directory to cache generated zero-knowledge proofs",
	})

	provingScheme := cfg.P2P.ZKP.ProvingScheme
	if provingScheme == "" {
		provingScheme = "plonk"
	}
	form.AddField(&tuicore.Field{
		Key: "zkp_proving_scheme", Label: "Proving Scheme", Type: tuicore.InputSelect,
		Value:       provingScheme,
		Options:     []string{"plonk", "groth16"},
		Description: "ZKP proving system: plonk=universal setup, groth16=faster but circuit-specific",
	})

	srsMode := cfg.P2P.ZKP.SRSMode
	if srsMode == "" {
		srsMode = "unsafe"
	}
	form.AddField(&tuicore.Field{
		Key: "zkp_srs_mode", Label: "SRS Mode", Type: tuicore.InputSelect,
		Value:       srsMode,
		Options:     []string{"unsafe", "file"},
		Description: "Structured Reference String mode: unsafe=dev-only random, file=from trusted setup",
	})

	form.AddField(&tuicore.Field{
		Key: "zkp_srs_path", Label: "SRS File Path", Type: tuicore.InputText,
		Value:       cfg.P2P.ZKP.SRSPath,
		Placeholder: "/path/to/srs.bin (when SRS mode = file)",
		Description: "Path to the SRS file from a trusted ceremony (required when mode=file)",
	})

	form.AddField(&tuicore.Field{
		Key: "zkp_max_credential_age", Label: "Max Credential Age", Type: tuicore.InputText,
		Value:       cfg.P2P.ZKP.MaxCredentialAge,
		Placeholder: "24h",
		Description: "Maximum age of a ZKP credential before it must be refreshed",
	})

	return &form
}

// NewP2PPricingForm creates the P2P Pricing configuration form.
func NewP2PPricingForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Pricing Configuration")

	form.AddField(&tuicore.Field{
		Key: "pricing_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.P2P.Pricing.Enabled,
		Description: "Enable paid tool invocations from P2P peers",
	})

	form.AddField(&tuicore.Field{
		Key: "pricing_per_query", Label: "Price Per Query (USDC)", Type: tuicore.InputText,
		Value:       cfg.P2P.Pricing.PerQuery,
		Placeholder: "0.50",
		Description: "USDC price charged per incoming P2P query",
	})

	form.AddField(&tuicore.Field{
		Key: "pricing_tool_prices", Label: "Tool Prices", Type: tuicore.InputText,
		Value:       formatKeyValueMap(cfg.P2P.Pricing.ToolPrices),
		Placeholder: "exec:0.10,browser:0.50 (name:price, comma-sep)",
		Description: "Per-tool USDC pricing overrides in tool_name:price format",
	})

	return &form
}

// NewP2POwnerProtectionForm creates the P2P Owner Protection configuration form.
func NewP2POwnerProtectionForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Owner Protection")

	form.AddField(&tuicore.Field{
		Key: "owner_name", Label: "Owner Name", Type: tuicore.InputText,
		Value:       cfg.P2P.OwnerProtection.OwnerName,
		Placeholder: "Your name to block from P2P responses",
		Description: "Owner's real name to prevent leaking via P2P responses",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_email", Label: "Owner Email", Type: tuicore.InputText,
		Value:       cfg.P2P.OwnerProtection.OwnerEmail,
		Placeholder: "your@email.com",
		Description: "Owner's email address to redact from P2P responses",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_phone", Label: "Owner Phone", Type: tuicore.InputText,
		Value:       cfg.P2P.OwnerProtection.OwnerPhone,
		Placeholder: "+82-10-1234-5678",
		Description: "Owner's phone number to redact from P2P responses",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_extra_terms", Label: "Extra Terms", Type: tuicore.InputText,
		Value:       strings.Join(cfg.P2P.OwnerProtection.ExtraTerms, ","),
		Placeholder: "company-name,project-name (comma-sep)",
		Description: "Additional terms to block from P2P responses (company names, etc.)",
	})

	form.AddField(&tuicore.Field{
		Key: "owner_block_conversations", Label: "Block Conversations", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.P2P.OwnerProtection.BlockConversations, true),
		Description: "Block P2P peers from accessing owner's conversation history",
	})

	return &form
}

// NewP2PSandboxForm creates the P2P Sandbox configuration form.
func NewP2PSandboxForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Sandbox Configuration")

	form.AddField(&tuicore.Field{
		Key: "sandbox_enabled", Label: "Tool Isolation Enabled", Type: tuicore.InputBool,
		Checked:     cfg.P2P.ToolIsolation.Enabled,
		Description: "Isolate P2P tool executions in sandboxed environments",
	})

	form.AddField(&tuicore.Field{
		Key: "sandbox_timeout", Label: "Timeout Per Tool", Type: tuicore.InputText,
		Value:       cfg.P2P.ToolIsolation.TimeoutPerTool.String(),
		Placeholder: "30s",
		Description: "Maximum execution time for a single sandboxed tool invocation",
	})

	form.AddField(&tuicore.Field{
		Key: "sandbox_max_memory_mb", Label: "Max Memory (MB)", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.P2P.ToolIsolation.MaxMemoryMB),
		Placeholder: "256",
		Description: "Memory limit in MB for each sandboxed tool execution",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	containerEnabled := &tuicore.Field{
		Key: "container_enabled", Label: "Container Sandbox", Type: tuicore.InputBool,
		Checked:     cfg.P2P.ToolIsolation.Container.Enabled,
		Description: "Use container-based isolation (Docker/gVisor) for stronger security",
	}
	form.AddField(containerEnabled)
	isContainerOn := func() bool { return containerEnabled.Checked }

	runtime := cfg.P2P.ToolIsolation.Container.Runtime
	if runtime == "" {
		runtime = "auto"
	}
	form.AddField(&tuicore.Field{
		Key: "container_runtime", Label: "  Runtime", Type: tuicore.InputSelect,
		Value:       runtime,
		Options:     []string{"auto", "docker", "gvisor", "native"},
		Description: "Container runtime: auto=detect best, gvisor=strongest isolation",
		VisibleWhen: isContainerOn,
	})

	form.AddField(&tuicore.Field{
		Key: "container_image", Label: "  Image", Type: tuicore.InputText,
		Value:       cfg.P2P.ToolIsolation.Container.Image,
		Placeholder: "lango-sandbox:latest",
		Description: "Docker image to use for sandboxed tool execution",
		VisibleWhen: isContainerOn,
	})

	networkMode := cfg.P2P.ToolIsolation.Container.NetworkMode
	if networkMode == "" {
		networkMode = "none"
	}
	form.AddField(&tuicore.Field{
		Key: "container_network_mode", Label: "  Network Mode", Type: tuicore.InputSelect,
		Value:       networkMode,
		Options:     []string{"none", "host", "bridge"},
		Description: "Container network: none=no network, host=full access, bridge=isolated",
		VisibleWhen: isContainerOn,
	})

	form.AddField(&tuicore.Field{
		Key: "container_readonly_rootfs", Label: "  Read-Only Rootfs", Type: tuicore.InputBool,
		Checked:     derefBool(cfg.P2P.ToolIsolation.Container.ReadOnlyRootfs, true),
		Description: "Mount container root filesystem as read-only for security",
		VisibleWhen: isContainerOn,
	})

	form.AddField(&tuicore.Field{
		Key: "container_cpu_quota", Label: "  CPU Quota (us)", Type: tuicore.InputInt,
		Value:       strconv.FormatInt(cfg.P2P.ToolIsolation.Container.CPUQuotaUS, 10),
		Placeholder: "0 (0 = unlimited)",
		Description: "CPU quota in microseconds per 100ms period; 0 = unlimited",
		VisibleWhen: isContainerOn,
		Validate: func(s string) error {
			if i, err := strconv.ParseInt(s, 10, 64); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "container_pool_size", Label: "  Pool Size", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.P2P.ToolIsolation.Container.PoolSize),
		Placeholder: "0 (0 = disabled)",
		Description: "Number of pre-warmed containers in the pool; 0 = create on demand",
		VisibleWhen: isContainerOn,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "container_pool_idle_timeout", Label: "  Pool Idle Timeout", Type: tuicore.InputText,
		Value:       cfg.P2P.ToolIsolation.Container.PoolIdleTimeout.String(),
		Placeholder: "5m",
		Description: "Time before idle pooled containers are destroyed",
		VisibleWhen: isContainerOn,
	})

	return &form
}

// NewP2PWorkspaceForm creates the P2P Workspace configuration form.
func NewP2PWorkspaceForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("P2P Workspace Configuration")

	wsEnabled := &tuicore.Field{
		Key: "ws_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.P2P.Workspace.Enabled,
		Description: "Enable collaborative workspaces for multi-agent co-work",
	}
	form.AddField(wsEnabled)
	isWSEnabled := func() bool { return wsEnabled.Checked }

	form.AddField(&tuicore.Field{
		Key: "ws_data_dir", Label: "Data Directory", Type: tuicore.InputText,
		Value:       cfg.P2P.Workspace.DataDir,
		Placeholder: "~/.lango/workspaces",
		Description: "Directory for storing workspace data and git bundles",
	})

	form.AddField(&tuicore.Field{
		Key: "ws_max_workspaces", Label: "Max Workspaces", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.P2P.Workspace.MaxWorkspaces),
		Description: "Maximum number of concurrent active workspaces",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "ws_max_bundle_size", Label: "Max Bundle Size (bytes)", Type: tuicore.InputInt,
		Value:       strconv.FormatInt(cfg.P2P.Workspace.MaxBundleSizeBytes, 10),
		Description: "Maximum git bundle size in bytes (0 = unlimited)",
		Validate: func(s string) error {
			if i, err := strconv.ParseInt(s, 10, 64); err != nil || i < 0 {
				return fmt.Errorf("must be a non-negative integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "ws_chronicler", Label: "Chronicler", Type: tuicore.InputBool,
		Checked:     cfg.P2P.Workspace.ChroniclerEnabled,
		Description: "Record workspace activity as triples in the graph store",
		VisibleWhen: isWSEnabled,
	})

	form.AddField(&tuicore.Field{
		Key: "ws_auto_sandbox", Label: "Auto Sandbox", Type: tuicore.InputBool,
		Checked:     cfg.P2P.Workspace.AutoSandbox,
		Description: "Automatically sandbox workspace operations for isolation",
		VisibleWhen: isWSEnabled,
	})

	form.AddField(&tuicore.Field{
		Key: "ws_contribution_tracking", Label: "Contribution Tracking", Type: tuicore.InputBool,
		Checked:     cfg.P2P.Workspace.ContributionTracking,
		Description: "Track per-agent contribution metrics (commits, messages, code bytes)",
		VisibleWhen: isWSEnabled,
	})

	return &form
}
