package config

import "time"

// SandboxConfig defines general-purpose OS-level tool execution sandbox settings.
// This applies to child processes spawned by exec tools, MCP stdio servers, and skill scripts.
// It is independent of p2p.toolIsolation, which controls subprocess/container isolation for P2P.
type SandboxConfig struct {
	// Enabled turns on OS-level sandboxing for tool-spawned child processes.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// FailClosed rejects tool execution when OS sandbox is unavailable (default: false = fail-open).
	FailClosed bool `mapstructure:"failClosed" json:"failClosed"`

	// Backend selects the OS sandbox isolation backend: "auto" (default), "seatbelt", "bwrap", "native", "none".
	// "auto" probes available backends and selects the best one.
	// "seatbelt" is macOS only.
	// "bwrap" requires the bubblewrap binary on Linux.
	// "native" (Landlock+seccomp) is not yet implemented.
	// "none" disables OS isolation even when sandbox.enabled is true.
	// Invalid values are rejected at startup.
	Backend string `mapstructure:"backend" json:"backend"`

	// WorkspacePath is the root directory for workspace-relative write access.
	// Defaults to CWD when empty.
	WorkspacePath string `mapstructure:"workspacePath" json:"workspacePath,omitempty"`

	// NetworkMode controls network access from sandboxed processes: "deny" or "allow" (default: "deny").
	// On Linux, enforced via bwrap when backend=bwrap or backend=auto:
	//   "deny"  → bwrap --unshare-net (new network namespace, lo down)
	//   "allow" → host network (no namespace unshare)
	NetworkMode string `mapstructure:"networkMode" json:"networkMode"`

	// AllowedNetworkIPs are IP addresses permitted for outbound connections (macOS Seatbelt only).
	// On Linux/bwrap this field is ignored — bwrap has no AF_INET filter, only
	// the all-or-nothing --unshare-net flag.
	AllowedNetworkIPs []string `mapstructure:"allowedNetworkIPs" json:"allowedNetworkIPs,omitempty"`

	// AllowedWritePaths are additional paths writable from the sandbox (beyond WorkspacePath).
	AllowedWritePaths []string `mapstructure:"allowedWritePaths" json:"allowedWritePaths,omitempty"`

	// TimeoutPerTool is the maximum duration for a single sandboxed tool execution (default: 30s).
	TimeoutPerTool time.Duration `mapstructure:"timeoutPerTool" json:"timeoutPerTool,omitempty"`

	// OS holds platform-specific sandbox settings.
	OS OSSandboxConfig `mapstructure:"os" json:"os"`
}

// OSSandboxConfig holds platform-specific sandbox settings.
type OSSandboxConfig struct {
	// SeccompProfile selects the seccomp filter profile on Linux: "strict", "moderate", or "permissive".
	// Default: "moderate". NOT YET ENFORCED — the native (Landlock+seccomp)
	// backend is planned. The bwrap backend does not consume this field.
	SeccompProfile string `mapstructure:"seccompProfile" json:"seccompProfile,omitempty"`

	// SeatbeltCustomProfile is a path to a custom .sb profile on macOS (overrides generated profile).
	SeatbeltCustomProfile string `mapstructure:"seatbeltCustomProfile" json:"seatbeltCustomProfile,omitempty"`
}
