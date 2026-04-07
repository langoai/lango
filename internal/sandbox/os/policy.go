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
// Read-global, write-workspace+/tmp, no network. .git inside the workspace and
// the entire lango control-plane (dataRoot, e.g. ~/.lango) are denied so that
// sandboxed children cannot read or write the agent's own state, secrets,
// session database, skills directory, or other internal data.
//
// Pass an empty dataRoot to skip the control-plane mask (used in unit tests
// that probe the policy in isolation).
func DefaultToolPolicy(workDir, dataRoot string) Policy {
	workDir, _ = filepath.Abs(workDir)
	denyPaths := []string{filepath.Join(workDir, ".git")}
	if dataRoot != "" {
		abs, err := filepath.Abs(dataRoot)
		if err == nil {
			denyPaths = append(denyPaths, abs)
		}
	}
	return Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			WritePaths:     []string{workDir, "/tmp"},
			DenyPaths:      denyPaths,
		},
		Network: NetworkDeny,
		Process: ProcessPolicy{
			AllowFork:    true,
			AllowSignals: false,
		},
	}
}

// StrictToolPolicy returns a maximally restrictive policy. Currently identical
// to DefaultToolPolicy because .git denial and control-plane masking are now
// part of the baseline. The function is preserved as a separate symbol so
// callers (and future strict-only options) can branch later without another
// signature migration.
func StrictToolPolicy(workDir, dataRoot string) Policy {
	return DefaultToolPolicy(workDir, dataRoot)
}

// MCPServerPolicy returns a policy for MCP stdio server processes.
// Read-global, write-/tmp only, network allowed (MCP servers need network).
// The lango control-plane (dataRoot) is denied so that misbehaving MCP server
// child processes cannot read or write lango's internal state.
//
// Pass an empty dataRoot to skip the control-plane mask (used in unit tests).
func MCPServerPolicy(dataRoot string) Policy {
	var denyPaths []string
	if dataRoot != "" {
		abs, err := filepath.Abs(dataRoot)
		if err == nil {
			denyPaths = []string{abs}
		}
	}
	return Policy{
		Filesystem: FilesystemPolicy{
			ReadOnlyGlobal: true,
			WritePaths:     []string{"/tmp"},
			DenyPaths:      denyPaths,
		},
		Network: NetworkAllow,
		Process: ProcessPolicy{
			AllowFork:    true,
			AllowSignals: false,
		},
	}
}
