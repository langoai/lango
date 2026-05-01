## MODIFIED Requirements

### Requirement: DynamicAllowedTools with runtime essentials

`DynamicAllowedTools` blocks SHALL be classified against the teammate role max scope before approval is requested. The capability layer SHALL request approval only for in-scope tools, SHALL deny out-of-scope tools directly, and SHALL extend the run's `AllowedTools` projection exactly once after approval so the next run context includes the granted tool.

#### Scenario: In-scope blocked tool requests approval
- **GIVEN** a teammate role whose max scope includes `fs_write`
- **AND** the current dynamic allowlist does not include `fs_write`
- **WHEN** the teammate attempts to call `fs_write`
- **THEN** the capability layer SHALL classify the block as in-scope
- **AND** SHALL request approval instead of denying the tool permanently

#### Scenario: Out-of-scope blocked tool is denied
- **GIVEN** a librarian teammate whose role max scope does not include `exec_shell`
- **WHEN** the teammate attempts to call `exec_shell`
- **THEN** the capability layer SHALL deny the request without approval

#### Scenario: Approved tool is added exactly once
- **GIVEN** approval is granted for an in-scope blocked tool
- **WHEN** the runtime updates the run projection
- **THEN** the granted tool SHALL be appended to `AllowedTools` exactly once
- **AND** the next execution context SHALL include the granted tool

#### Scenario: Runtime essentials remain allowed
- **GIVEN** a context with `DynamicAllowedTools` set to a narrowed allowlist
- **WHEN** the teammate calls `builtin_list`
- **THEN** runtime essential handling SHALL continue to allow the call
