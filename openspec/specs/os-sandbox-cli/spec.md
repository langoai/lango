## ADDED Requirements

### Requirement: sandbox status command output
`lango sandbox status` SHALL display: Sandbox Configuration (enabled, fail-mode explanation when enabled, network mode, workspace), Active Isolation (isolator name, available, reason if unavailable), and Platform Capabilities (platform, kernel, each primitive with reason-aware status). The capability formatter SHALL use reason strings from `PlatformCapabilities` instead of bool-only display. Primitives with `"probe not yet implemented"` reason SHALL display as `"unknown (probe not yet implemented)"` rather than `"unavailable"`.

#### Scenario: Linux status with noop isolator
- **WHEN** `lango sandbox status` runs on Linux with no isolation backend
- **THEN** output shows `Isolator: noop`, `Reason: Linux isolation backend not yet implemented`, and `Landlock: unknown (probe not yet implemented)`

#### Scenario: macOS status with seatbelt
- **WHEN** `lango sandbox status` runs on macOS with sandbox-exec available
- **THEN** output shows `Isolator: seatbelt`, `Available: true`, and `Seatbelt: available (sandbox-exec found)`

#### Scenario: Fail-mode display
- **WHEN** sandbox is enabled with `failClosed=false`
- **THEN** status shows `Fail-Closed: fail-open (warning + unsandboxed execution)`

#### Scenario: Status shows allowedNetworkIPs warning on Linux
- **WHEN** `lango sandbox status` is run on Linux with `allowedNetworkIPs` configured
- **THEN** output SHALL include a warning that `allowedNetworkIPs` is macOS-only

### Requirement: TUI settings descriptions
The OS Sandbox settings form and menu descriptions SHALL accurately reflect Linux enforcement status. Descriptions SHALL NOT claim Linux Landlock/seccomp enforcement when it is not implemented.

#### Scenario: Form description accuracy
- **WHEN** user views OS Sandbox settings form
- **THEN** enabled field description says "Seatbelt on macOS; Linux: planned, not yet enforced"
- **AND** seccomp profile field description says "Linux only — not yet enforced"
- **AND** menu description says "OS-level tool isolation (macOS enforced, Linux planned)"

### Requirement: Sandbox test command
The system SHALL provide `lango sandbox test` that runs smoke tests verifying filesystem write restriction and read permission.

#### Scenario: Test on platform with sandbox
- **WHEN** `lango sandbox test` is run with available sandbox
- **THEN** it SHALL verify write to /etc is blocked and read from /etc/hosts succeeds

#### Scenario: Test on platform without sandbox
- **WHEN** `lango sandbox test` is run without available sandbox
- **THEN** it SHALL print "OS sandbox not available on this platform"
