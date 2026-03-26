## ADDED Requirements

### Requirement: OS Sandbox settings form
The settings TUI SHALL provide an "OS Sandbox" category under the Security section with 9 fields mapping to `cfg.Sandbox.*`, using `os_sandbox_*` field key prefix.

#### Scenario: Form contains all sandbox config fields
- **WHEN** `NewOSSandboxForm(cfg)` is called
- **THEN** the form SHALL contain 9 fields: os_sandbox_enabled, os_sandbox_fail_closed, os_sandbox_workspace_path, os_sandbox_network_mode, os_sandbox_allowed_ips, os_sandbox_allowed_write_paths, os_sandbox_timeout, os_sandbox_seccomp_profile, os_sandbox_seatbelt_profile

#### Scenario: OS sandbox fields do not affect P2P sandbox config
- **WHEN** `os_sandbox_enabled` is toggled in the form
- **THEN** `cfg.Sandbox.Enabled` SHALL change and `cfg.P2P.ToolIsolation.Enabled` SHALL NOT change

#### Scenario: Menu includes OS Sandbox category
- **WHEN** the settings menu is rendered
- **THEN** the Security section SHALL contain an "OS Sandbox" entry with ID `os_sandbox`

#### Scenario: OS Sandbox category is always enabled
- **WHEN** `categoryIsEnabled("os_sandbox")` is called
- **THEN** it SHALL return true regardless of other config settings
