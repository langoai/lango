## MODIFIED Requirements

### Requirement: Config set uses single bootstrap with cleanup
The `config set` command SHALL bootstrap exactly once. The cfgLoader function SHALL return a cleanup function that closes the DB client. The cleanup function MUST be called via `defer` in `RunE` to ensure resources are released on all code paths (success, setConfigPath error, save error).

#### Scenario: Config get/set/keys output uses the command writer
- **WHEN** `lango config get`, `lango config set`, or `lango config keys` renders output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output
