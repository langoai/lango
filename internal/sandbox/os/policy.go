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

// findGitRoot walks upward from workDir looking for the first ancestor that
// contains a `.git` directory. Returns the absolute path to that `.git`
// directory (e.g. "/home/user/repo/.git"), or "" if no git root is found
// before reaching the filesystem root. Called from the policy builders so
// that a workDir which happens to be a subdirectory of a git repo (e.g.
// supervisor cwd = /repo/cmd/lango while .git lives at /repo/.git) still
// gets the baseline .git deny applied.
//
// Worktree pointers (.git as a regular file) are skipped — walk-up keeps
// climbing past them because compileBwrapArgs cannot mount --tmpfs on a
// file. File-level deny semantics will arrive with PR 5c, which closes
// the worktree gap.
//
// The walk terminates when filepath.Dir(cur)==cur (reached filesystem
// root). An empty or unresolvable workDir also returns "" — callers
// simply drop the .git baseline and continue, which matches the
// "non-repo workspace" trade-off.
func findGitRoot(workDir string) string {
	if workDir == "" {
		return ""
	}
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return ""
	}
	cur := abs
	for {
		if candidate := filepath.Join(cur, ".git"); isDir(candidate) {
			return candidate
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return ""
		}
		cur = parent
	}
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
// Read-global, write-workspace+/tmp, no network. The first ancestor `.git`
// directory discovered via upward walk from workDir and the entire lango
// control-plane (dataRoot, e.g. ~/.lango) are denied so that sandboxed
// children cannot read or write the agent's own state, secrets, session
// database, skills directory, or other internal data.
//
// Baseline deny paths are added only when they exist as directories. When
// workDir is a subdirectory of a git repo, findGitRoot walks up to locate
// the repo root `.git`, so callers that pass cwd (not the repo root) still
// get correct protection. A fully non-repo workspace or a worktree pointer
// (`.git` as a regular file) is silently skipped — file-level deny arrives
// in PR 5c. A missing dataRoot is also skipped; pass an empty dataRoot to
// intentionally drop the mask.
func DefaultToolPolicy(workDir, dataRoot string) Policy {
	workDir, _ = filepath.Abs(workDir)
	var denyPaths []string
	if gitDir := findGitRoot(workDir); gitDir != "" {
		denyPaths = append(denyPaths, gitDir)
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
// The first ancestor `.git` directory discovered via walk-up from workDir
// and the lango control-plane (dataRoot) are denied so that misbehaving MCP
// server child processes cannot read or write git metadata or lango's
// internal state. This mirrors DefaultToolPolicy's baseline deny — both
// sandboxed-children surfaces share the same protection.
//
// Baseline deny paths are added only when they exist as directories. An
// empty workDir or a workDir with no ancestor `.git` directory is silently
// skipped. A missing or non-directory dataRoot is also skipped so that
// isolated unit tests and minimal environments can still build the policy.
// Pass empty strings to intentionally drop the masks.
func MCPServerPolicy(workDir, dataRoot string) Policy {
	var denyPaths []string
	if workDir != "" {
		if gitDir := findGitRoot(workDir); gitDir != "" {
			denyPaths = append(denyPaths, gitDir)
		}
	}
	if dataRoot != "" {
		if abs, err := filepath.Abs(dataRoot); err == nil && isDir(abs) {
			denyPaths = append(denyPaths, abs)
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
