package os

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// normalizePath runs the canonical sandbox-policy path normalization pipeline.
// All backends (bwrap, Seatbelt, planned native) use this helper so every
// consumer sees entries in the same shape:
//
//	entry → sanitize → Abs → Glob → EvalSymlinks (with fallback) → []string
//
// Step-by-step:
//  1. sanitizePath runs filepath.Abs and rejects injection characters (`"`,
//     `(`, `)`, `;`, newline). Returns ErrInvalidPolicy on violation.
//  2. If the sanitized path contains a wildcard (`*`, `?`, `[`), filepath.Glob
//     expands it. Zero matches → silent empty slice (shell nullglob semantics,
//     aligning with the existing best-effort baseline deny philosophy).
//     An invalid glob pattern (filepath.ErrBadPattern, unclosed bracket, etc.)
//     surfaces as an error — user config mistakes should fail loudly at
//     startup, not be silently dropped.
//  3. For each glob match (or the single literal path), filepath.EvalSymlinks
//     resolves the path through any symlink chain. On os.IsNotExist, fall back
//     to the pre-resolve absolute path so downstream os.Stat catches the
//     missing-path error with the existing message format. Other EvalSymlinks
//     errors (permission denied walking the chain) are wrapped and returned.
//
// Returns zero or more concrete paths (nil on silent skip, non-empty on
// match) or an error from any pipeline stage. Callers append each returned
// path to their backend-specific argv / profile.
func normalizePath(entry string) ([]string, error) {
	absPath, err := sanitizePath(entry)
	if err != nil {
		return nil, err
	}

	var candidates []string
	if strings.ContainsAny(absPath, "*?[") {
		matches, err := filepath.Glob(absPath)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %q: %w", entry, err)
		}
		if len(matches) == 0 {
			// Silent skip — shell nullglob semantics. Debug logging is the
			// caller's responsibility if they want an audit trail.
			return nil, nil
		}
		candidates = matches
	} else {
		candidates = []string{absPath}
	}

	results := make([]string, 0, len(candidates))
	for _, c := range candidates {
		resolved, err := filepath.EvalSymlinks(c)
		if err != nil {
			if os.IsNotExist(err) {
				// Missing target — let the downstream os.Stat catch it with
				// the existing error format.
				results = append(results, c)
				continue
			}
			return nil, fmt.Errorf("resolve symlinks %q: %w", c, err)
		}
		results = append(results, resolved)
	}
	return results, nil
}

// gitRoot describes the result of findGitRoot — a pair of paths identifying
// the git metadata location discovered via upward walk from a workDir.
//
//   - pointerPath is the discovered `.git` entry itself. For a standard
//     repository this is the `.git` directory. For a linked worktree it is
//     the `.git` file that contains a `gitdir: <path>` pointer.
//   - gitdirPath is the resolved gitdir target. For a directory, it equals
//     pointerPath. For a worktree pointer file whose contents were parsed
//     successfully, it is the absolute path to the target gitdir (which may
//     lie outside the workspace — that is normal for worktrees created via
//     `git worktree add`). If the pointer file is malformed or unreadable,
//     gitdirPath is empty and callers should fall back to denying only the
//     pointer file itself (degraded protection — see PR 5c design doc).
//
// A zero value (both fields empty) means no git metadata was found.
type gitRoot struct {
	pointerPath string
	gitdirPath  string
}

// found reports whether the walk discovered any git metadata.
func (r gitRoot) found() bool { return r.pointerPath != "" }

// findGitRoot walks upward from workDir looking for the first ancestor whose
// `.git` entry is a directory (standard repo) or a regular file with a
// `gitdir: <path>` pointer (linked worktree). Symlinked workDirs are
// resolved via filepath.EvalSymlinks before the walk so that the canonical
// filesystem path is used throughout.
//
// Walk-up terminates when filepath.Dir(cur)==cur (reached filesystem root).
// Empty or unresolvable workDir returns the zero gitRoot.
//
// For worktree pointers, the gitdir target is resolved via the same
// canonical pipeline (Abs + EvalSymlinks). Relative gitdir paths are
// resolved against the directory containing the pointer file. Malformed or
// unreadable pointers degrade to pointer-only denial (gitdirPath is empty)
// — callers still deny the pointer file via file-level deny (PR 5c) which
// at least blocks direct reads of the gitdir pointer.
//
// The walk continues past `.git` entries that are neither directory nor
// regular file (device nodes, sockets, fifos — extremely unusual, likely
// user filesystem corruption). This preserves the existing silent-skip
// philosophy for pathological cases.
func findGitRoot(workDir string) gitRoot {
	if workDir == "" {
		return gitRoot{}
	}
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return gitRoot{}
	}
	// Resolve workDir symlinks so the walk operates on the canonical path.
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	cur := abs
	for {
		candidate := filepath.Join(cur, ".git")
		if fi, err := os.Stat(candidate); err == nil {
			switch {
			case fi.IsDir():
				return gitRoot{pointerPath: candidate, gitdirPath: candidate}
			case fi.Mode().IsRegular():
				return parseWorktreePointer(candidate)
			default:
				// Device/socket/fifo — skip, walk-up continues.
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return gitRoot{}
		}
		cur = parent
	}
}

// parseWorktreePointer reads a `.git` regular file expected to contain a
// `gitdir: <path>` line and resolves the target. Relative gitdir paths are
// interpreted against the directory containing the pointer file. Absolute
// paths and resolved targets both flow through filepath.EvalSymlinks.
//
// On any read/parse/resolve failure the function returns a degraded gitRoot
// with pointerPath set but gitdirPath empty — callers still get file-level
// denial of the pointer file itself, which at least blocks direct reads.
func parseWorktreePointer(pointerPath string) gitRoot {
	data, err := os.ReadFile(pointerPath)
	if err != nil {
		return gitRoot{pointerPath: pointerPath}
	}
	// Only the first line matters; git writes exactly "gitdir: <path>\n".
	firstLine := strings.SplitN(strings.TrimSpace(string(data)), "\n", 2)[0]
	const prefix = "gitdir:"
	if !strings.HasPrefix(firstLine, prefix) {
		return gitRoot{pointerPath: pointerPath}
	}
	target := strings.TrimSpace(firstLine[len(prefix):])
	if target == "" {
		return gitRoot{pointerPath: pointerPath}
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(pointerPath), target)
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return gitRoot{pointerPath: pointerPath}
	}
	if resolved, err := filepath.EvalSymlinks(absTarget); err == nil {
		absTarget = resolved
	}
	return gitRoot{pointerPath: pointerPath, gitdirPath: absTarget}
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
// Read-global, write-workspace+/tmp, no network. The first ancestor git
// metadata discovered via upward walk from workDir and the entire lango
// control-plane (dataRoot, e.g. ~/.lango) are denied so that sandboxed
// children cannot read or write the agent's own state, secrets, session
// database, skills directory, or other internal data.
//
// Git metadata discovery uses findGitRoot's two-path model: a standard
// `.git` directory adds one DenyPaths entry; a linked worktree `.git` file
// adds two — the pointer file itself (denied at file level via PR 5c) AND
// the resolved gitdir target it points to. A fully non-repo workspace or a
// worktree pointer whose content is malformed silently degrades to
// whatever coverage is possible. A missing dataRoot is also skipped; pass
// an empty dataRoot to intentionally drop the mask.
//
// workDir itself flows through filepath.Abs + filepath.EvalSymlinks so
// that WritePaths[0] is the canonical filesystem path (symlinked workDirs
// no longer leak their pre-resolve path into the writable set).
func DefaultToolPolicy(workDir, dataRoot string) Policy {
	return DefaultToolPolicyWithProtectedPaths(workDir, dataRoot, nil)
}

// DefaultToolPolicyWithProtectedPaths augments the baseline deny set with
// additional resolved runtime protected paths.
func DefaultToolPolicyWithProtectedPaths(workDir, dataRoot string, protectedPaths []string) Policy {
	workDir = canonicalWorkDir(workDir)
	denyPaths := collectBaselineDeny(workDir, dataRoot, protectedPaths)
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

// canonicalWorkDir resolves workDir through filepath.Abs + EvalSymlinks so
// that all derived paths (WritePaths, git walk-up starting point) operate
// on the canonical filesystem path. Missing paths fall back to the Abs
// result — downstream consumers catch nonexistent workDirs at Stat time
// with the existing error format.
func canonicalWorkDir(workDir string) string {
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return workDir
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}

// collectBaselineDeny builds the DenyPaths slice for the default tool and
// MCP policies. It applies the two-path gitRoot model (pointer + gitdir
// target, distinct paths when they differ) and adds the resolved dataRoot
// when non-empty and existing.
func collectBaselineDeny(workDir, dataRoot string, protectedPaths []string) []string {
	var denyPaths []string
	if workDir != "" {
		if gr := findGitRoot(workDir); gr.found() {
			denyPaths = append(denyPaths, gr.pointerPath)
			if gr.gitdirPath != "" && gr.gitdirPath != gr.pointerPath {
				denyPaths = append(denyPaths, gr.gitdirPath)
			}
		}
	}
	if dataRoot != "" {
		if abs, err := filepath.Abs(dataRoot); err == nil && isDir(abs) {
			denyPaths = append(denyPaths, abs)
		}
	}
	for _, path := range protectedPaths {
		if path == "" {
			continue
		}
		if abs, err := filepath.Abs(path); err == nil {
			denyPaths = append(denyPaths, abs)
		}
	}
	return denyPaths
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
	return MCPServerPolicyWithProtectedPaths(workDir, dataRoot, nil)
}

// MCPServerPolicyWithProtectedPaths augments the baseline deny set with
// additional resolved runtime protected paths.
func MCPServerPolicyWithProtectedPaths(workDir, dataRoot string, protectedPaths []string) Policy {
	if workDir != "" {
		workDir = canonicalWorkDir(workDir)
	}
	denyPaths := collectBaselineDeny(workDir, dataRoot, protectedPaths)
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
