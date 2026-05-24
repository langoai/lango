## MODIFIED Requirements

### Requirement: DynamicAllowedTools with runtime essentials

`DynamicAllowedTools` remains the runtime ceiling for tool execution. For built-in teammate types, that ceiling is interpreted under the teammate role maximum scope. For custom or non-built-in teammate paths, the current allowlist itself SHALL be treated as the effective ceiling because no built-in role registry applies.

#### Scenario: Custom teammate uses current allowlist as its ceiling
- **GIVEN** a non-built-in teammate with `CurrentAllowed: ["exec"]`
- **WHEN** capability policy evaluates tool `exec`
- **THEN** the runtime SHALL treat `exec` as in-scope because it is already present in the current allowlist
- **AND** the runtime SHALL NOT require a built-in teammate role definition for that decision

#### Scenario: Custom teammate cannot escalate beyond current allowlist
- **GIVEN** a non-built-in teammate with `CurrentAllowed: ["fs_read"]`
- **WHEN** capability policy evaluates tool `exec`
- **THEN** the runtime SHALL treat `exec` as out-of-scope
- **AND** the current allowlist SHALL act as the effective ceiling for that teammate path
