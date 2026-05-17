## MODIFIED Requirements

### Requirement: sandbox status command output
`lango sandbox status` SHALL display: Sandbox Configuration (enabled, fail-mode explanation when enabled and not opted out, backend label, network mode, workspace), Active Isolation (isolator name, available, reason if unavailable), Platform Capabilities (platform, kernel, primitives), and **Backend Availability** (one row per platform candidate with available/unavailable status and reason). The capability formatter SHALL distinguish between `"unknown (probe not yet implemented)"` and `"unavailable (reason)"`.

`lango sandbox status` SHALL accept `--output table|json|plain`. The default SHALL be `table` and preserve the existing multi-section human-readable output. `json` SHALL emit a single JSON object containing configuration, active isolation, platform capabilities, backend availability, Linux warning state, and recent sandbox decisions when available. `plain` SHALL emit concise key-value text without table section headers.

Invalid output formats SHALL fail before loading config or bootstrap state.

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

#### Scenario: macOS status with seatbelt
- **WHEN** `lango sandbox status` runs on macOS with sandbox-exec available
- **THEN** output shows `Isolator: seatbelt`, `Available: true`, and `Seatbelt: available (sandbox-exec found)`

#### Scenario: Fail-mode display
- **WHEN** sandbox is enabled with `failClosed=false` and not opted out
- **THEN** status shows `Fail-Closed: fail-open (warning + unsandboxed execution)`

#### Scenario: Status shows allowedNetworkIPs warning on Linux
- **WHEN** `lango sandbox status` is run on Linux with `allowedNetworkIPs` configured
- **THEN** output SHALL include a warning that `allowedNetworkIPs` is macOS-only

#### Scenario: JSON output is scriptable
- **WHEN** `lango sandbox status --output json` runs
- **THEN** stdout SHALL be valid JSON
- **AND** the JSON SHALL include `configuration`, `active_isolation`, `platform_capabilities`, `backend_availability`, and `recent_decisions`

#### Scenario: Plain output is concise
- **WHEN** `lango sandbox status --output plain` runs
- **THEN** stdout SHALL contain key-value lines for enabled state, backend, isolator, platform, workspace, and backend availability
- **AND** stdout SHALL NOT contain table section headers such as `Sandbox Configuration:`

#### Scenario: Invalid output rejected before loading
- **WHEN** `lango sandbox status --output yaml` runs
- **THEN** the command SHALL fail before invoking the config loader or bootstrap loader
