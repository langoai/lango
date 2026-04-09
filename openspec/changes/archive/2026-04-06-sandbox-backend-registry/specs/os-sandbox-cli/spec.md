## MODIFIED Requirements

### Requirement: sandbox status command output
`lango sandbox status` SHALL display: Sandbox Configuration (enabled, fail-mode explanation when enabled and not opted out, backend label, network mode, workspace), Active Isolation (isolator name, available, reason if unavailable), Platform Capabilities (platform, kernel, primitives), and **Backend Availability** (one row per platform candidate with available/unavailable status and reason). The capability formatter SHALL distinguish between `"unknown (probe not yet implemented)"` and `"unavailable (reason)"`.

When `sandbox.enabled=true` and `sandbox.backend=none`, status SHALL display `"Backend: none (explicit opt-out — fail-closed not applied)"` and SHALL NOT print the `Fail-Closed` line, accurately reflecting that the runtime skips fail-closed for this configuration.

#### Scenario: Backend Availability section present
- **WHEN** `lango sandbox status` runs
- **THEN** output contains a `Backend Availability:` header followed by one row per platform candidate using `ListBackends(PlatformBackendCandidates())`

#### Scenario: Auto resolved label
- **WHEN** `sandbox.backend=auto` and seatbelt is selected
- **THEN** status shows `"Backend: auto (resolved: seatbelt)"`

#### Scenario: backend=none opt-out display
- **WHEN** `sandbox.enabled=true` and `sandbox.backend=none`
- **THEN** status shows `"Backend: none (explicit opt-out — fail-closed not applied)"` and omits the Fail-Closed line

#### Scenario: Linux status with noop isolator
- **WHEN** `lango sandbox status` runs on Linux with no isolation backend
- **THEN** output shows `Isolator: noop` and the noop's `Reason()` field aggregates each candidate's reason

### Requirement: sandbox test command honors configured backend
`lango sandbox test` SHALL accept a `cfgLoader` callback and use `ParseBackendMode(cfg.Sandbox.Backend) + SelectBackend(mode, PlatformBackendCandidates())` instead of `NewOSIsolator()`. When the configured backend is `none`, test SHALL print a message indicating no isolation to test and exit successfully without running smoke tests. When the configured backend is unavailable, test SHALL print the backend name and reason and exit successfully.

#### Scenario: backend=none short-circuits test
- **WHEN** `sandbox.backend=none` and `lango sandbox test` runs
- **THEN** output contains `"no isolation to test"` and the command exits successfully without running write/read tests

#### Scenario: Unavailable backend reports reason
- **WHEN** `sandbox.backend=bwrap` (stub, unavailable) and `lango sandbox test` runs
- **THEN** output contains `"Sandbox backend bwrap not available"` and the bwrap reason

### Requirement: TUI settings form exposes backend selection
The OS Sandbox settings form SHALL include an `os_sandbox_backend` field of type `InputSelect` with options `["auto", "seatbelt", "bwrap", "native", "none"]`. The TUI state-update layer SHALL map this field to `cfg.Sandbox.Backend`.

#### Scenario: Backend select field present
- **WHEN** the OS Sandbox form is rendered
- **THEN** it contains a select field keyed `os_sandbox_backend` with the five backend options
