package os

import (
	"os"
	"path/filepath"
)

// isDir reports whether p exists and is a directory. Used by the policy
// builders to guard baseline deny paths (.git, dataRoot) against missing or
// non-directory entries, which would otherwise cause compileBwrapArgs on
// Linux to reject the entire policy (bwrap --tmpfs requires an existing
// directory). Missing entries are silently skipped; the caller treats the
// absent deny as a trade-off (see design note: worktree .git is a file).
func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

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
// Baseline deny paths are added only when they exist as directories. A missing
// workDir/.git (non-repo workspace) or a .git file (linked worktree) is
// silently skipped so the policy remains buildable. A missing dataRoot is
// also skipped — pass an empty dataRoot to intentionally drop the mask.
func DefaultToolPolicy(workDir, dataRoot string) Policy {
	workDir, _ = filepath.Abs(workDir)
	var denyPaths []string
	if gitPath := filepath.Join(workDir, ".git"); isDir(gitPath) {
		denyPaths = append(denyPaths, gitPath)
	}
	if dataRoot != "" {
		if abs, err := filepath.Abs(dataRoot); err == nil && isDir(abs) {
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
// The dataRoot deny is added only when dataRoot exists as a directory. A
// missing or non-directory dataRoot is silently skipped so that isolated
// unit tests and minimal environments can still build the policy. Pass an
// empty dataRoot to intentionally drop the mask.
func MCPServerPolicy(dataRoot string) Policy {
	var denyPaths []string
	if dataRoot != "" {
		if abs, err := filepath.Abs(dataRoot); err == nil && isDir(abs) {
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
