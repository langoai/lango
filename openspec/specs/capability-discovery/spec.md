# capability-discovery Specification

## Purpose
TBD - created by archiving change ux-capability-concierge. Update Purpose after archive.
## Requirements
### Requirement: list_skills summary parameter
The `list_skills` tool SHALL accept an optional `summary` boolean parameter (default `false`). When `summary=true`, the tool SHALL return only `{name, description, when_to_use}` for each active skill, omitting full content and reference paths. When `summary=false` (default), the tool SHALL return its existing full output, preserving backward compatibility.

#### Scenario: Summary mode returns metadata only
- **WHEN** `list_skills` is called with `summary=true`
- **THEN** the result SHALL contain only `name`, `description`, and `when_to_use` fields per skill

#### Scenario: Default full output preserved
- **WHEN** `list_skills` is called without the `summary` parameter
- **THEN** the output SHALL match the existing full format

#### Scenario: Mode-filtered summary
- **WHEN** the session has an active mode with a `Skills` allowlist and `list_skills(summary=true)` is invoked
- **THEN** only skills in the mode's allowlist SHALL appear in the result

### Requirement: view_skill tool
The system SHALL provide a `view_skill` tool with parameters `name string` (required) and `path string` (optional). With `name` only, the tool SHALL return the full SKILL.md for that skill. With `name` and `path`, the tool SHALL return the content of the referenced supporting file, resolved relative to the skill's directory. Requests for paths outside the skill directory SHALL return an error.

#### Scenario: view_skill returns full SKILL.md
- **WHEN** `view_skill(name="pytest-runner")` is called for an active skill
- **THEN** the full SKILL.md content SHALL be returned

#### Scenario: view_skill returns reference file
- **WHEN** `view_skill(name="pytest-runner", path="references/api.md")` is called
- **THEN** the content of `<skills-dir>/pytest-runner/references/api.md` SHALL be returned

#### Scenario: Path escape rejected
- **WHEN** `view_skill(name="pytest-runner", path="../../../../etc/passwd")` is called
- **THEN** the tool SHALL return an error indicating the path is outside the skill directory

#### Scenario: Unknown skill rejected
- **WHEN** `view_skill(name="nonexistent")` is called
- **THEN** the tool SHALL return an error indicating the skill is not active

### Requirement: Deferred exposure for instruction and template skills
Skills with type `instruction` or `template` SHALL register with `ExposureDeferred` so they do NOT appear as individual tools in the system prompt. They remain discoverable via `list_skills` and loadable via `view_skill`. Skills with type `script` or `fork` SHALL retain direct-call exposure.

#### Scenario: Instruction skill not in tool catalog
- **WHEN** a skill has `type: instruction` and the system prompt is assembled
- **THEN** the skill SHALL NOT appear as a standalone tool description in the prompt

#### Scenario: Script skill remains directly callable
- **WHEN** a skill has `type: script`
- **THEN** the skill SHALL register as a directly callable tool with normal exposure

### Requirement: Capability discovery prompt guidance
When a session has an active mode, the system prompt SHALL include guidance instructing the LLM to use `list_skills(summary=true)` to scan available skills and `view_skill(name)` to load full content on demand. Without an active mode, existing behavior SHALL be preserved.

#### Scenario: Mode adds capability discovery hint
- **WHEN** a session with an active mode generates its system prompt
- **THEN** the prompt SHALL contain guidance referencing `list_skills(summary=true)` and `view_skill`

