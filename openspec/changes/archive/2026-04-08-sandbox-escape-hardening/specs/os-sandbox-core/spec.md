## MODIFIED Requirements

### Requirement: Policy types
The system SHALL define `Policy` with `FilesystemPolicy` (ReadOnlyGlobal, ReadPaths, WritePaths, DenyPaths), `NetworkPolicy` (deny/allow/unix-only), `ProcessPolicy` (AllowFork, AllowSignals), and `AllowedNetworkIPs` (macOS only).

The policy helper functions SHALL accept a `dataRoot` parameter so they can deny the lango control-plane (typically `~/.lango`) on every sandboxed child. An empty `dataRoot` SHALL skip the control-plane mask so isolated unit tests can build a policy without fabricating a real directory.

`DefaultToolPolicy(workDir, dataRoot string) Policy` SHALL deny `<workDir>/.git` as a baseline (previously a `StrictToolPolicy`-only feature). When `dataRoot` is non-empty, it SHALL be resolved via `filepath.Abs` and added to `DenyPaths` so the lango data directory is masked from the sandboxed child.

`StrictToolPolicy(workDir, dataRoot string) Policy` SHALL currently return the same policy as `DefaultToolPolicy`. The function is preserved as a separate symbol so future strict-only options can branch later without another signature migration.

`MCPServerPolicy(dataRoot string) Policy` SHALL deny `dataRoot` (when non-empty) so MCP stdio server child processes cannot read or write the lango control-plane. It retains its other characteristics: read-global, write-`/tmp` only, network allowed.

#### Scenario: Default tool policy denies dataRoot and .git
- **WHEN** `DefaultToolPolicy("/home/user/project", "/home/user/.lango")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain `/home/user/project/.git` and `/home/user/.lango`
- **AND** `policy.Filesystem.WritePaths` SHALL contain `/home/user/project` and `/tmp`
- **AND** `policy.Network` SHALL equal `NetworkDeny`

#### Scenario: Default tool policy with empty dataRoot
- **WHEN** `DefaultToolPolicy("/home/user/project", "")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain only `/home/user/project/.git`
- **AND** the policy SHALL be safe to use in isolated unit tests that do not have a real control-plane directory

#### Scenario: Strict tool policy mirrors default
- **WHEN** `StrictToolPolicy(workDir, dataRoot)` is called with the same arguments as `DefaultToolPolicy(workDir, dataRoot)`
- **THEN** the returned policies SHALL be equal (including DenyPaths order)

#### Scenario: MCP server policy denies dataRoot
- **WHEN** `MCPServerPolicy("/home/user/.lango")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL contain `/home/user/.lango`
- **AND** `policy.Network` SHALL equal `NetworkAllow`

#### Scenario: MCP server policy with empty dataRoot
- **WHEN** `MCPServerPolicy("")` is called
- **THEN** `policy.Filesystem.DenyPaths` SHALL be empty

### Requirement: Seatbelt profile generation
The system SHALL generate macOS Seatbelt `.sb` profiles from Policy via `text/template` with default-deny base, path sanitization against injection characters, and IP allowlist support.

When `Policy.Filesystem.DenyPaths` contains an entry, the generated profile SHALL include a `(deny file-write* (subpath "<path>"))` rule for that entry. The control-plane deny added by `DefaultToolPolicy(workDir, dataRoot)` SHALL therefore appear in the generated Seatbelt profile when `dataRoot` is non-empty.

#### Scenario: Profile blocks injection characters
- **WHEN** a path contains `"`, `(`, `)`, or `;`
- **THEN** `GenerateSeatbeltProfile()` SHALL return `ErrInvalidPolicy`

#### Scenario: Profile includes allowed IPs
- **WHEN** Policy has `AllowedNetworkIPs` with entries
- **THEN** the profile SHALL contain `(allow network-outbound (remote ip "..."))` rules

#### Scenario: Profile denies dataRoot when configured
- **WHEN** `GenerateSeatbeltProfile(DefaultToolPolicy("/tmp/work", "/home/user/.lango"))` is called
- **THEN** the profile SHALL contain `(deny file-write* (subpath "/home/user/.lango"))`
