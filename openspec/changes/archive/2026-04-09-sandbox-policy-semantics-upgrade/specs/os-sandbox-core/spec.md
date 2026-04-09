## MODIFIED Requirements

### Requirement: Policy types
The system SHALL define `Policy` with `FilesystemPolicy` (ReadOnlyGlobal, ReadPaths, WritePaths, DenyPaths), `NetworkPolicy` (deny/allow/unix-only), `ProcessPolicy` (AllowFork, AllowSignals), and `AllowedNetworkIPs` (macOS only).

The policy helper functions SHALL accept a `dataRoot` parameter so they can deny the lango control-plane (typically `~/.lango`) on every sandboxed child. An empty `dataRoot` SHALL skip the control-plane mask so isolated unit tests can build a policy without fabricating a real directory.

`DefaultToolPolicy(workDir, dataRoot string) Policy` SHALL route `workDir` through a `canonicalWorkDir` helper (`filepath.Abs` + `filepath.EvalSymlinks` with nonexistent fallback) so that `WritePaths[0]` is the canonical filesystem path — symlinked workspaces no longer leak their pre-resolve path into the writable set. It SHALL then call a private `findGitRoot` helper that walks upward from the canonical workDir looking for the first ancestor whose `.git` entry is a directory (standard repo) or a regular file (linked worktree pointer). The walk terminates when `filepath.Dir(cur) == cur` (filesystem root).

`findGitRoot` SHALL return a `gitRoot` struct with `pointerPath` and `gitdirPath` fields. For a standard `.git` directory, both fields equal the `.git` directory path. For a linked worktree `.git` file, `pointerPath` is the file itself and `gitdirPath` is the resolved target parsed from the file's `gitdir: <path>` line (relative targets resolved against the pointer file's parent directory, then flowed through `filepath.Abs` + `filepath.EvalSymlinks`). Malformed or unreadable pointer files SHALL degrade to `gitdirPath = ""` — callers still deny the pointer file itself via file-level deny.

`DefaultToolPolicy` and `MCPServerPolicy` SHALL share a `collectBaselineDeny(workDir, dataRoot)` helper that applies the two-deny strategy: the gitRoot `pointerPath` and (when non-empty and distinct) `gitdirPath` are both added to `DenyPaths`, followed by the resolved `dataRoot` (when non-empty and existing as a directory). This means:
- Standard repo → one DenyPaths entry (`.git` directory)
- Linked worktree → two DenyPaths entries (pointer file + gitdir target, which may lie outside the workspace — that is the normal worktree layout)
- Malformed worktree pointer → one DenyPaths entry (pointer file only, degraded)
- Non-repo workspace → zero entries from git; dataRoot added when present

`StrictToolPolicy(workDir, dataRoot string) Policy` SHALL currently return the same policy as `DefaultToolPolicy`.

`MCPServerPolicy(workDir, dataRoot string) Policy` SHALL apply the same `collectBaselineDeny` logic via the shared helper so MCP stdio server children get symmetric protection with exec/skill tools. Empty `workDir` or `dataRoot` SHALL silently skip the corresponding baseline deny.

The system SHALL also provide a `normalizePath(entry string) ([]string, error)` helper in `internal/sandbox/os/policy.go` that runs the canonical path normalization pipeline shared by all sandbox backends:

```
entry → sanitize → filepath.Abs → filepath.Glob → filepath.EvalSymlinks (with nonexistent fallback) → []string
```

Every policy path consumer (bwrap `compileBwrapArgs` ReadPaths/WritePaths/DenyPaths loops; Seatbelt `GenerateSeatbeltProfile` ReadPaths/WritePaths/DenyPaths loops; future native Linux backend) SHALL use `normalizePath` so bwrap and Seatbelt cannot drift. Zero-match globs silently skip; invalid glob patterns return an error; nonexistent paths fall back to the pre-resolve absolute path so downstream `os.Stat` catches the missing-path error with the existing error message format.

#### Scenario: Default tool policy denies existing .git directory and dataRoot
- **WHEN** `DefaultToolPolicy(workDir, dataRoot)` is called AND an ancestor `.git` directory exists (immediate or via walk-up) AND `dataRoot` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain both the discovered `.git` path AND `dataRoot`
- **AND** `policy.Filesystem.WritePaths` SHALL contain the canonical workDir path (symlink-resolved) and `/tmp` (symlink-resolved on macOS to `/private/tmp`)
- **AND** `policy.Network` SHALL equal `NetworkDeny`

#### Scenario: Default tool policy walks up to find ancestor .git
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `workDir` is a subdirectory of a git repository whose `.git` directory exists at an ancestor path
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain the absolute path of the discovered ancestor `.git` directory, not a fictional path under `workDir`

#### Scenario: No ancestor .git is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND no ancestor directory of `workDir` contains a `.git` directory (non-repo workspace)
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain any `.git` path
- **AND** the policy SHALL still be buildable

#### Scenario: Linked worktree denies both pointer and gitdir target
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `<workDir>/.git` exists as a regular file containing `gitdir: <target>` AND the target resolves to an existing directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain BOTH the `.git` pointer file AND the resolved gitdir target (two entries)
- **AND** the pointer file deny flows through file-level bwrap/Seatbelt enforcement (`--ro-bind /dev/null` for bwrap)

#### Scenario: Malformed worktree pointer degrades to pointer-only deny
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `<workDir>/.git` exists as a regular file that does NOT start with `gitdir:`
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain the `.git` pointer file (one entry)
- **AND** the policy SHALL still be buildable with degraded protection

#### Scenario: Symlinked workDir resolves to canonical path
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `workDir` is a symlink to a real directory
- **THEN** `policy.Filesystem.WritePaths[0]` SHALL equal the symlink-resolved real path, NOT the original symlink path
- **AND** `findGitRoot` SHALL walk up from the resolved real path

#### Scenario: Missing dataRoot is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, dataRoot)` is called AND `dataRoot` is non-empty but does not exist as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain `dataRoot`
- **AND** the policy SHALL still be buildable

#### Scenario: Strict tool policy mirrors default
- **WHEN** `StrictToolPolicy(workDir, dataRoot)` is called with the same arguments as `DefaultToolPolicy(workDir, dataRoot)`
- **THEN** the returned policies SHALL be equal (including DenyPaths order)

#### Scenario: MCP server policy denies existing dataRoot
- **WHEN** `MCPServerPolicy("", dataRoot)` is called AND `dataRoot` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain `dataRoot`
- **AND** `policy.Network` SHALL equal `NetworkAllow`

#### Scenario: MCP server policy with empty workDir and empty dataRoot
- **WHEN** `MCPServerPolicy("", "")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL be empty
- **AND** `policy.Network` SHALL equal `NetworkAllow`

#### Scenario: MCP server policy with missing dataRoot
- **WHEN** `MCPServerPolicy("", dataRoot)` is called AND `dataRoot` is non-empty but does not exist as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL be empty

#### Scenario: MCP server policy denies workspace .git via walk-up
- **WHEN** `MCPServerPolicy(workDir, "")` is called AND `workDir` is non-empty AND an ancestor `.git` directory exists via walk-up
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain the discovered ancestor `.git` path (symmetric with `DefaultToolPolicy`)

#### Scenario: MCP server policy denies both workspace .git and dataRoot
- **WHEN** `MCPServerPolicy(workDir, dataRoot)` is called AND an ancestor `.git` exists AND `dataRoot` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain both the discovered `.git` path AND `dataRoot`

#### Scenario: normalizePath pipeline expands globs
- **WHEN** `normalizePath(pattern)` is called with a pattern containing `*`, `?`, or `[` AND the pattern matches one or more files
- **THEN** it SHALL return a slice with one entry per match, each run through `filepath.EvalSymlinks`

#### Scenario: normalizePath silently skips unmatched globs
- **WHEN** `normalizePath(pattern)` is called with a pattern that matches zero files
- **THEN** it SHALL return a nil slice with a nil error

#### Scenario: normalizePath rejects invalid glob patterns
- **WHEN** `normalizePath(entry)` is called with an entry containing an unclosed bracket
- **THEN** it SHALL return an error containing `"invalid glob pattern"`

#### Scenario: normalizePath falls back to Abs path for nonexistent entries
- **WHEN** `normalizePath(entry)` is called with a nonexistent path that has no glob characters
- **THEN** it SHALL return a slice containing the `filepath.Abs` result (not an error) so downstream `os.Stat` catches the missing-path error with the existing format
