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

	// WorkspacePath is the root directory for workspace-relative write access.
	// Defaults to CWD when empty.
	WorkspacePath string `mapstructure:"workspacePath" json:"workspacePath,omitempty"`

	// NetworkMode controls network access from sandboxed processes: "deny" or "allow" (default: "deny").
	// On Linux, this setting is not yet enforced (isolation backend planned).
	NetworkMode string `mapstructure:"networkMode" json:"networkMode"`

	// AllowedNetworkIPs are IP addresses permitted for outbound connections (macOS Seatbelt only).
	// On Linux, this field is ignored — Linux isolation is not yet enforced.
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
	// Default: "moderate".
	SeccompProfile string `mapstructure:"seccompProfile" json:"seccompProfile,omitempty"`

	// SeatbeltCustomProfile is a path to a custom .sb profile on macOS (overrides generated profile).
	SeatbeltCustomProfile string `mapstructure:"seatbeltCustomProfile" json:"seatbeltCustomProfile,omitempty"`
}
