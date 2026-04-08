## MODIFIED Requirements

### Requirement: Policy types
The system SHALL define `Policy` with `FilesystemPolicy` (ReadOnlyGlobal, ReadPaths, WritePaths, DenyPaths), `NetworkPolicy` (deny/allow/unix-only), `ProcessPolicy` (AllowFork, AllowSignals), and `AllowedNetworkIPs` (macOS only).

The policy helper functions SHALL accept a `dataRoot` parameter so they can deny the lango control-plane (typically `~/.lango`) on every sandboxed child. An empty `dataRoot` SHALL skip the control-plane mask so isolated unit tests can build a policy without fabricating a real directory.

`DefaultToolPolicy(workDir, dataRoot string) Policy` SHALL walk up from `workDir` via a private `findGitRoot` helper to discover the first ancestor `.git` directory and deny it as a baseline. `findGitRoot` SHALL terminate when `filepath.Dir(cur) == cur` (filesystem root) and SHALL return an empty string if no ancestor `.git` directory is found or if the only ancestor `.git` entry is a regular file (linked worktree pointer). When `dataRoot` is non-empty, it SHALL be resolved via `filepath.Abs` and added to `DenyPaths` only when the resolved path exists as a directory. Missing or non-directory entries SHALL be silently skipped so the policy remains buildable in non-repo workspaces, linked worktrees (where `.git` is a file and walk-up finds no ancestor directory), and minimal environments where the control-plane directory has not been created yet. This gate exists because `compileBwrapArgs` requires every deny path to exist as a directory and would otherwise fail the entire sandbox apply.

`StrictToolPolicy(workDir, dataRoot string) Policy` SHALL currently return the same policy as `DefaultToolPolicy`. The function is preserved as a separate symbol so future strict-only options can branch later without another signature migration.

`MCPServerPolicy(workDir, dataRoot string) Policy` SHALL apply the same walk-up `.git` baseline deny as `DefaultToolPolicy` via `findGitRoot(workDir)` when `workDir` is non-empty. It SHALL also deny `dataRoot` (when non-empty and when the resolved path exists as a directory) so MCP stdio server child processes cannot read or write the lango control-plane. Empty `workDir` or `dataRoot` SHALL silently skip the corresponding baseline deny. Other characteristics are retained: read-global, write-`/tmp` only, network allowed.

#### Scenario: Default tool policy denies existing .git and dataRoot
- **WHEN** `DefaultToolPolicy(workDir, dataRoot)` is called AND an ancestor `.git` directory exists (immediate or via walk-up) AND `dataRoot` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain both the discovered `.git` path AND `dataRoot`
- **AND** `policy.Filesystem.WritePaths` SHALL contain `workDir` and `/tmp`
- **AND** `policy.Network` SHALL equal `NetworkDeny`

#### Scenario: Default tool policy walks up to find ancestor .git
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `workDir` is a subdirectory of a git repository whose `.git` directory exists at an ancestor path (e.g. `workDir=/repo/cmd/lango`, ancestor `.git=/repo/.git`)
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain the absolute path of the discovered ancestor `.git` directory, not a fictional path under `workDir`

#### Scenario: No ancestor .git is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND no ancestor directory of `workDir` contains a `.git` directory (non-repo workspace)
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain any `.git` path
- **AND** the policy SHALL still be buildable so that `compileBwrapArgs` does not reject non-repo workspaces

#### Scenario: .git file (worktree) causes walk-up to continue past it
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `<workDir>/.git` exists as a regular file (linked worktree pointer) AND no ancestor directory above contains a `.git` directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain the `.git` file path
- **AND** the policy SHALL still be buildable
- **AND** walk-up SHALL NOT return the worktree file as a match (this gap closes in PR 5c with file-level deny semantics)

#### Scenario: Missing dataRoot is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, dataRoot)` is called AND `dataRoot` is non-empty but does not exist as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain `dataRoot`
- **AND** the policy SHALL still be buildable

#### Scenario: Default tool policy with empty dataRoot
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND an ancestor `.git` directory exists
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain only the discovered ancestor `.git` path
- **AND** the policy SHALL be safe to use in isolated unit tests that do not have a real control-plane directory

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
