## MODIFIED Requirements

### Requirement: Policy types
The system SHALL define `Policy` with `FilesystemPolicy` (ReadOnlyGlobal, ReadPaths, WritePaths, DenyPaths), `NetworkPolicy` (deny/allow/unix-only), `ProcessPolicy` (AllowFork, AllowSignals), and `AllowedNetworkIPs` (macOS only).

The policy helper functions SHALL accept a `dataRoot` parameter so they can deny the lango control-plane (typically `~/.lango`) on every sandboxed child. An empty `dataRoot` SHALL skip the control-plane mask so isolated unit tests can build a policy without fabricating a real directory.

`DefaultToolPolicy(workDir, dataRoot string) Policy` SHALL deny `<workDir>/.git` as a baseline when the path exists as a directory. When `dataRoot` is non-empty, it SHALL be resolved via `filepath.Abs` and added to `DenyPaths` only when the resolved path exists as a directory. Missing or non-directory entries SHALL be silently skipped so the policy remains buildable in non-repo workspaces, linked worktrees (where `.git` is a file), and minimal environments where the control-plane directory has not been created yet. This gate exists because `compileBwrapArgs` requires every deny path to exist as a directory and would otherwise fail the entire sandbox apply.

`StrictToolPolicy(workDir, dataRoot string) Policy` SHALL currently return the same policy as `DefaultToolPolicy`. The function is preserved as a separate symbol so future strict-only options can branch later without another signature migration.

`MCPServerPolicy(dataRoot string) Policy` SHALL deny `dataRoot` (when non-empty and when the resolved path exists as a directory) so MCP stdio server child processes cannot read or write the lango control-plane. It retains its other characteristics: read-global, write-`/tmp` only, network allowed.

#### Scenario: Default tool policy denies existing .git and dataRoot
- **WHEN** `DefaultToolPolicy(workDir, dataRoot)` is called AND `<workDir>/.git` exists as a directory AND `dataRoot` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain both paths
- **AND** `policy.Filesystem.WritePaths` SHALL contain `workDir` and `/tmp`
- **AND** `policy.Network` SHALL equal `NetworkDeny`

#### Scenario: Missing .git is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `<workDir>/.git` does not exist
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain the `.git` path
- **AND** the policy SHALL still be buildable so that `compileBwrapArgs` does not reject non-repo workspaces

#### Scenario: .git file (worktree) is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `<workDir>/.git` exists as a regular file (linked worktree)
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain the `.git` file path
- **AND** the policy SHALL still be buildable

#### Scenario: Missing dataRoot is silently skipped
- **WHEN** `DefaultToolPolicy(workDir, dataRoot)` is called AND `dataRoot` is non-empty but does not exist as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL NOT contain `dataRoot`
- **AND** the policy SHALL still be buildable

#### Scenario: Default tool policy with empty dataRoot
- **WHEN** `DefaultToolPolicy(workDir, "")` is called AND `<workDir>/.git` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain only `<workDir>/.git`
- **AND** the policy SHALL be safe to use in isolated unit tests that do not have a real control-plane directory

#### Scenario: Strict tool policy mirrors default
- **WHEN** `StrictToolPolicy(workDir, dataRoot)` is called with the same arguments as `DefaultToolPolicy(workDir, dataRoot)`
- **THEN** the returned policies SHALL be equal (including DenyPaths order)

#### Scenario: MCP server policy denies existing dataRoot
- **WHEN** `MCPServerPolicy(dataRoot)` is called AND `dataRoot` exists as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain `dataRoot`
- **AND** `policy.Network` SHALL equal `NetworkAllow`

#### Scenario: MCP server policy with empty dataRoot
- **WHEN** `MCPServerPolicy("")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL be empty

#### Scenario: MCP server policy with missing dataRoot
- **WHEN** `MCPServerPolicy(dataRoot)` is called AND `dataRoot` is non-empty but does not exist as a directory
- **THEN** `policy.Filesystem.DenyPaths` SHALL be empty

### Requirement: Seatbelt profile generation
The system SHALL generate macOS Seatbelt `.sb` profiles from Policy via `text/template` with default-deny base, path sanitization against injection characters, and IP allowlist support.

When `Policy.Filesystem.DenyPaths` contains an entry, the generated profile SHALL include BOTH a `(deny file-read* (subpath "<path>"))` rule AND a `(deny file-write* (subpath "<path>"))` rule for that entry. The read-deny rule is required because `ReadOnlyGlobal=true` emits `(allow file-read*)` globally, and a write-only deny would leave sensitive files (such as `~/.lango/lango.db`, `.git/config`, or encrypted config tokens) fully readable by the sandboxed child. The control-plane deny added by `DefaultToolPolicy(workDir, dataRoot)` SHALL therefore appear in the generated Seatbelt profile as both read-deny and write-deny rules when `dataRoot` is non-empty and exists.

#### Scenario: Profile blocks injection characters
- **WHEN** a path contains `"`, `(`, `)`, or `;`
- **THEN** `GenerateSeatbeltProfile()` SHALL return `ErrInvalidPolicy`

#### Scenario: Profile includes allowed IPs
- **WHEN** Policy has `AllowedNetworkIPs` with entries
- **THEN** the profile SHALL contain `(allow network-outbound (remote ip "..."))` rules

#### Scenario: DenyPaths entries deny both read and write
- **WHEN** `Policy.Filesystem.DenyPaths` contains `/home/user/.lango` and `ReadOnlyGlobal=true`
- **THEN** the generated profile SHALL contain BOTH `(deny file-read* (subpath "/home/user/.lango"))` AND `(deny file-write* (subpath "/home/user/.lango"))`
- **AND** the sandboxed child SHALL NOT be able to read `/home/user/.lango/lango.db` or any other file under the control-plane directory

#### Scenario: Profile denies dataRoot for read and write when configured
- **WHEN** `GenerateSeatbeltProfile` is called on a Policy whose DenyPaths contains `/home/user/.lango`
- **THEN** the profile SHALL contain `(deny file-read* (subpath "/home/user/.lango"))` AND `(deny file-write* (subpath "/home/user/.lango"))`
