package os

import "path/filepath"

// NetworkPolicy controls network access from the sandbox.
type NetworkPolicy string

const (
	// NetworkDeny blocks all network access.
	NetworkDeny NetworkPolicy = "deny"

	// NetworkAllow permits unrestricted network access.
	NetworkAllow NetworkPolicy = "allow"

	// NetworkUnixOnly allows only AF_UNIX sockets (for local IPC/proxy).
	NetworkUnixOnly NetworkPolicy = "unix-only"
)

// FilesystemPolicy defines filesystem access rules.
type FilesystemPolicy struct {
	// ReadOnlyGlobal allows read access to the entire filesystem.
	// When true, ReadPaths is ignored — all paths are readable.
	ReadOnlyGlobal bool

	// ReadPaths are paths allowed for reading (used when ReadOnlyGlobal is false).
	ReadPaths []string

	// WritePaths are paths allowed for writing.
	WritePaths []string

	// DenyPaths are paths explicitly denied (takes precedence over allow).
	DenyPaths []string
}

// ProcessPolicy defines process creation controls.
type ProcessPolicy struct {
	// AllowFork permits fork/exec within the sandbox.
	AllowFork bool

	// AllowSignals permits sending signals to other processes.
	AllowSignals bool
}

// Policy defines the sandbox restrictions for a single tool execution.
type Policy struct {
	// Filesystem controls file access.
	Filesystem FilesystemPolicy

	// Network controls network access.
	Network NetworkPolicy

	// Process controls process creation and signals.
	Process ProcessPolicy

	// AllowedNetworkIPs are IP addresses allowed for outbound connections (macOS only).
	// On Linux, this field is ignored (seccomp cannot filter by IP).
	AllowedNetworkIPs []string
}

// DefaultToolPolicy returns the standard sandbox policy for local tool execution.
// Read-global, write-workspace+/tmp, no network. Matches Claude Code's approach.
func DefaultToolPolicy(workDir string) Policy {
	workDir, _ = filepath.Abs(workDir)
	return Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			WritePaths:     []string{workDir, "/tmp"},
		},
		Network: NetworkDeny,
		Process: ProcessPolicy{
			AllowFork:    true,
			AllowSignals: false,
		},
	}
}

// StrictToolPolicy returns a maximally restrictive policy.
// .git read-only, no network. Matches Codex CLI approach.
func StrictToolPolicy(workDir string) Policy {
	workDir, _ = filepath.Abs(workDir)
	return Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			WritePaths:     []string{workDir, "/tmp"},
			DenyPaths:      []string{filepath.Join(workDir, ".git")},
		},
		Network: NetworkDeny,
		Process: ProcessPolicy{
			AllowFork:    true,
			AllowSignals: false,
		},
	}
}

// MCPServerPolicy returns a policy for MCP stdio server processes.
// Read-global, write-/tmp only, network allowed (MCP servers need network).
func MCPServerPolicy() Policy {
	return Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			WritePaths:     []string{"/tmp"},
		},
		Network: NetworkAllow,
		Process: ProcessPolicy{
			AllowFork:    true,
			AllowSignals: false,
		},
	}
}
