## ADDED Requirements

### Requirement: Sandbox status graceful degradation
`lango sandbox status` SHALL render the Sandbox Configuration, Active Isolation, Platform Capabilities, and Backend Availability sections in degraded modes (signed-out, locked DB, non-interactive environments, or missing BootLoader) by falling back to a config-only loader. The Recent Sandbox Decisions section SHALL be silently skipped when the audit DB is unreachable, but the rest of the command SHALL NOT error out — these diagnostic sections do not depend on the audit database.

`newStatusCmd` SHALL try the BootLoader first so that one bootstrap pass serves both the config rendering and the Recent Decisions audit query (preserving the no-double-passphrase contract). On nil BootLoader OR a BootLoader error, `newStatusCmd` SHALL fall back to the cfgLoader to load the config independently. Recent Decisions SHALL only render when the BootLoader returned a non-nil result with a non-nil DBClient.

#### Scenario: Nil BootLoader still renders config sections
- **WHEN** `lango sandbox status` is invoked with cfgLoader wired and BootLoader nil
- **THEN** the command SHALL exit successfully
- **AND** the output SHALL contain the `Sandbox Configuration:`, `Active Isolation:`, and `Backend Availability:` headers
- **AND** the output SHALL NOT contain a `Recent Sandbox Decisions` header

#### Scenario: BootLoader error falls back to cfgLoader
- **WHEN** `lango sandbox status` is invoked with cfgLoader wired and BootLoader returning an error
- **THEN** the command SHALL exit successfully via the cfgLoader fallback
- **AND** the non-audit sections SHALL render
- **AND** the `Recent Sandbox Decisions` section SHALL be silently skipped

#### Scenario: Healthy BootLoader runs only one bootstrap
- **WHEN** `lango sandbox status` is invoked with both loaders wired and BootLoader succeeding
- **THEN** cfgLoader SHALL NOT be called (the Recent Decisions path uses `boot.Config` directly)
- **AND** the user SHALL be prompted for the encryption passphrase at most once per invocation

### Requirement: Sandbox decision row formatter
The `Recent Sandbox Decisions` row formatter SHALL display `-` in the backend column whenever the decision is NOT `"applied"` OR the stored backend value is empty. Only `"applied"` decisions actually ran inside a sandbox backend; `excluded`, `skipped`, and `rejected` verdicts ran unsandboxed (or were blocked entirely), so echoing the published `Backend` value for those rows would falsely suggest the command was sandboxed under that backend.

The publish sites (`exec`, `skill`, `mcp`) SHALL continue to stamp the `Backend` field uniformly from the wired isolator's `Name()` regardless of decision; the verdict-specific formatting is the display layer's responsibility, not the publisher's.

#### Scenario: Applied decision shows backend
- **WHEN** an audit row has `decision="applied"` and `backend="bwrap"`
- **THEN** the rendered row SHALL show `bwrap` in the backend column

#### Scenario: Excluded decision shows dash
- **WHEN** an audit row has `decision="excluded"` and `backend="bwrap"`
- **THEN** the rendered row SHALL show `-` in the backend column
- **AND** the rendered row SHALL NOT contain the substring `bwrap`

#### Scenario: Skipped decision shows dash
- **WHEN** an audit row has `decision="skipped"` and `backend="seatbelt"`
- **THEN** the rendered row SHALL show `-` in the backend column
- **AND** the rendered row SHALL NOT contain the substring `seatbelt`

#### Scenario: Rejected decision shows dash
- **WHEN** an audit row has `decision="rejected"` and `backend="bwrap"`
- **THEN** the rendered row SHALL show `-` in the backend column

#### Scenario: Empty backend shows dash
- **WHEN** an audit row has `decision="applied"` and `backend=""`
- **THEN** the rendered row SHALL show `-` in the backend column
