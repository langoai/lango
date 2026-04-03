# skill-runtime-v2 Specification

## Purpose
Skill Runtime v2 extends skills from storage format to execution context with conditional activation, model preferences, fork delegation, and project-local discovery. This spec covers the v2 enhancements to the `SkillEntry` schema, the `fork` skill type, path-conditional activation, project-local skill discovery, and SKILL.md frontmatter roundtrip for all v2 fields.

## Requirements

### Requirement: Extended SkillEntry schema
`SkillEntry` SHALL include the following additional fields beyond the base skill-system spec: `WhenToUse` (string), `Paths` ([]string), `Context` (string), `Model` (string), `Effort` (string), `Agent` (string), `Hooks` (map[string]string). All v2 fields SHALL be optional and default to their zero values.

#### Scenario: Full v2 field population
- **WHEN** a `SkillEntry` is created with all v2 fields set (WhenToUse, Paths, Context, Model, Effort, Agent, Hooks)
- **THEN** all fields SHALL be stored and retrievable on the entry

#### Scenario: Backward compatibility with empty v2 fields
- **WHEN** a legacy SKILL.md without any v2 fields is parsed
- **THEN** the resulting `SkillEntry` SHALL have all v2 fields at their zero values (empty strings, nil slices, nil maps)
- **AND** the skill SHALL function identically to pre-v2 behavior

#### Scenario: Partial v2 field population
- **WHEN** a `SkillEntry` has only some v2 fields set (e.g., WhenToUse and Effort)
- **THEN** the set fields SHALL be stored and the remaining fields SHALL remain at their zero values

### Requirement: SkillTypeFork
The system SHALL support a `fork` skill type (`SkillTypeFork = "fork"`) that returns model-guided delegation text directing the agent to transfer work to a specialist agent. The fork type SHALL be included in the valid skill types enumeration.

#### Scenario: Fork execution with agent
- **WHEN** a fork skill with `Agent: "core-developer"` and `Definition["instruction"]: "Follow the refactoring plan"` is executed
- **THEN** the executor SHALL return delegation text referencing the `core-developer` agent with the instruction
- **AND** the text SHALL include a `transfer_to_agent('core-developer')` directive

#### Scenario: Fork execution with default agent
- **WHEN** a fork skill with an empty `Agent` field is executed
- **THEN** the executor SHALL default to `"operator"` as the target agent name

#### Scenario: Fork missing instruction error
- **WHEN** a fork skill with an empty `Definition["instruction"]` is executed
- **THEN** the executor SHALL return an error indicating the fork skill is missing an instruction

#### Scenario: Fork AllowedTools advisory
- **WHEN** a fork skill with `AllowedTools: ["exec", "fs_read"]` is executed
- **THEN** the delegation text SHALL include an advisory tool restrictions section listing `exec, fs_read`
- **AND** the text SHALL note that tool restrictions are enforced only when using `agent_spawn`

#### Scenario: Fork AllowedTools empty
- **WHEN** a fork skill with no `AllowedTools` is executed
- **THEN** the advisory tool restrictions section SHALL show `none`

#### Scenario: Fork with parameters
- **WHEN** a fork skill is executed with parameters `{"target": "internal/core"}`
- **THEN** the delegation text SHALL include a Parameters section listing the provided key-value pairs

#### Scenario: Fork type validation
- **WHEN** `SkillTypeFork.Valid()` is called
- **THEN** it SHALL return `true`
- **AND** `SkillTypeFork` SHALL appear in `SkillType.Values()`

#### Scenario: Fork skill registration as tool
- **WHEN** a fork skill is loaded by the registry
- **THEN** it SHALL be registered as `skill_{name}` with capability category `skill` and activity `ActivityExecute`

#### Scenario: Fork skill default description
- **WHEN** a fork skill has an empty description
- **THEN** the tool description SHALL default to "Fork skill that delegates to the '{agent}' agent"

### Requirement: Path-conditional activation
The `Activator` SHALL match edited file paths against skill `Paths` globs using `filepath.Match`. A skill is considered matching if any of its glob patterns matches any of the edited paths.

#### Scenario: Exact filename match
- **WHEN** a skill has `Paths: ["main.go"]` and the edited path is `main.go`
- **THEN** the skill SHALL be returned by `CheckPaths`

#### Scenario: Glob star match
- **WHEN** a skill has `Paths: ["*.go"]` and the edited path is `handler.go`
- **THEN** the skill SHALL be returned by `CheckPaths`

#### Scenario: Directory glob match
- **WHEN** a skill has `Paths: ["cmd/*.go"]` and the edited path is `cmd/main.go`
- **THEN** the skill SHALL be returned by `CheckPaths`

#### Scenario: No match
- **WHEN** a skill has `Paths: ["*.py"]` and the edited path is `handler.go`
- **THEN** the skill SHALL NOT be returned by `CheckPaths`

#### Scenario: No edited paths
- **WHEN** `CheckPaths` is called with an empty or nil edited paths slice
- **THEN** it SHALL return nil without querying the registry

#### Scenario: Multiple skills, partial match
- **WHEN** multiple active skills exist, some with matching paths and some without
- **THEN** only the skills whose `Paths` match at least one edited path SHALL be returned

#### Scenario: Dedup across multiple globs
- **WHEN** a single skill has `Paths: ["*.go", "main.*"]` and the edited path is `main.go`
- **THEN** the skill SHALL appear exactly once in the result (not duplicated per matching glob)

#### Scenario: Multiple edited paths trigger
- **WHEN** a skill has `Paths: ["*_test.go"]` and edited paths include `main.go` and `handler_test.go`
- **THEN** the skill SHALL be returned (any match is sufficient)

#### Scenario: Skill with empty Paths is skipped
- **WHEN** a skill has `Paths: []` (empty slice)
- **THEN** it SHALL NOT be considered for path-based activation

#### Scenario: Malformed glob is skipped
- **WHEN** a skill has a malformed glob pattern (e.g., `[invalid`)
- **THEN** the malformed pattern SHALL be skipped gracefully without error

### Requirement: Project-local skill discovery
`FileSkillStore.DiscoverProjectSkills` SHALL scan `<projectRoot>/.lango/skills/` for directories containing `SKILL.md` files. Hidden directories (names starting with `.`) SHALL be skipped.

#### Scenario: Discover project skills
- **WHEN** `DiscoverProjectSkills` is called with a project root containing `.lango/skills/my-skill/SKILL.md`
- **THEN** the discovered skill SHALL be returned as a `SkillEntry`

#### Scenario: Missing skills directory
- **WHEN** `DiscoverProjectSkills` is called and `<projectRoot>/.lango/skills/` does not exist
- **THEN** it SHALL return nil with no error

#### Scenario: Hidden directory ignored
- **WHEN** `.lango/skills/` contains a directory starting with `.` (e.g., `.placeholder`)
- **THEN** that directory SHALL be skipped without attempting to parse its contents

#### Scenario: Invalid SKILL.md skipped
- **WHEN** a project skill directory contains a malformed SKILL.md
- **THEN** it SHALL be skipped with a warning log and not cause an error for other skills

### Requirement: Project-local skill name conflict resolution
When `Registry.LoadProjectSkills` merges project-local skills, name conflicts with already-loaded global skills SHALL be resolved in favor of the global skill.

#### Scenario: Global skill wins on conflict
- **WHEN** a global skill `skill_deploy` is already loaded and a project-local skill with the same name is discovered
- **THEN** the project-local skill SHALL be skipped with a warning log
- **AND** the global skill SHALL remain in the loaded tools

#### Scenario: No conflict
- **WHEN** a project-local skill has a name not present in the global skills
- **THEN** it SHALL be added to the loaded tools

### Requirement: SKILL.md frontmatter roundtrip
`ParseSkillMD` and `RenderSkillMD` SHALL roundtrip all v2 fields through YAML frontmatter. The `Paths` and `AllowedTools` fields SHALL be serialized as space-separated strings in the frontmatter and parsed back into `[]string` slices.

#### Scenario: Parse all v2 fields
- **WHEN** a SKILL.md contains `when_to_use`, `paths`, `context`, `model`, `effort`, `agent`, and `hooks` in its YAML frontmatter
- **THEN** `ParseSkillMD` SHALL populate the corresponding `SkillEntry` fields

#### Scenario: Render all v2 fields
- **WHEN** a `SkillEntry` with all v2 fields set is rendered via `RenderSkillMD`
- **THEN** the output SHALL contain all v2 fields in the YAML frontmatter

#### Scenario: Roundtrip preserves values
- **WHEN** a `SkillEntry` with v2 fields is rendered via `RenderSkillMD` and the output is parsed via `ParseSkillMD`
- **THEN** all v2 field values on the re-parsed entry SHALL equal the original values

#### Scenario: Empty v2 fields omitted in output
- **WHEN** a `SkillEntry` has all v2 fields at their zero values
- **THEN** `RenderSkillMD` SHALL NOT include `when_to_use`, `paths`, `context`, `model`, `effort`, `agent`, or `hooks` keys in the YAML frontmatter

#### Scenario: Partial v2 fields in output
- **WHEN** a `SkillEntry` has only `WhenToUse` and `Effort` set (other v2 fields empty)
- **THEN** `RenderSkillMD` SHALL include only `when_to_use` and `effort` in the frontmatter
- **AND** the other v2 keys SHALL NOT appear

#### Scenario: Paths roundtrip as space-separated string
- **WHEN** a `SkillEntry` has `Paths: ["src/**/*.go", "internal/**/*.go"]`
- **THEN** `RenderSkillMD` SHALL output `paths: src/**/*.go internal/**/*.go` in frontmatter
- **AND** re-parsing SHALL yield the original `[]string` slice

#### Scenario: Hooks roundtrip as map
- **WHEN** a `SkillEntry` has `Hooks: {"pre": "go vet ./...", "post": "go test ./..."}`
- **THEN** `RenderSkillMD` SHALL output the hooks as a YAML map
- **AND** re-parsing SHALL yield the original `map[string]string`

### Requirement: Fork body parsing
`ParseSkillMD` SHALL parse the body of a `fork` skill by storing the entire markdown body (excluding the `## Parameters` section) as `Definition["instruction"]`.

#### Scenario: Parse fork body
- **WHEN** a SKILL.md of type `fork` has a markdown body "Follow the refactoring plan carefully."
- **THEN** `ParseSkillMD` SHALL set `Definition["instruction"]` to the body content

#### Scenario: Fork body with parameters section
- **WHEN** a fork SKILL.md body contains a `## Parameters` section
- **THEN** the instruction SHALL exclude the parameters section
- **AND** the parameters SHALL be parsed separately into the `Parameters` field

### Requirement: Fork body rendering
`RenderSkillMD` SHALL render fork skills by outputting the instruction as plain markdown (no code block wrapping), consistent with instruction skill rendering.

#### Scenario: Render fork body
- **WHEN** a fork `SkillEntry` with `Definition["instruction"]: "Refactor Guide"` is rendered
- **THEN** the body SHALL contain `Refactor Guide` as plain text without code block delimiters

### Requirement: Default type inference
When `ParseSkillMD` encounters a SKILL.md with no explicit `type` in frontmatter, it SHALL default to `instruction` type.

#### Scenario: No type defaults to instruction
- **WHEN** a SKILL.md has no `type` field in frontmatter
- **THEN** the parsed `SkillEntry.Type` SHALL be `SkillTypeInstruction`

### Requirement: Default status inference
When `ParseSkillMD` encounters a SKILL.md with no explicit `status` in frontmatter, it SHALL default to `active` status.

#### Scenario: No status defaults to active
- **WHEN** a SKILL.md has no `status` field in frontmatter
- **THEN** the parsed `SkillEntry.Status` SHALL be `SkillStatusActive`
