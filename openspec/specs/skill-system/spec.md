## ADDED Requirements

### Requirement: File-Based Skill Storage
The system SHALL store skills as `<dir>/<name>/SKILL.md` files with YAML frontmatter containing name, description, type, status, and optional parameters. `ListActive()` SHALL skip hidden directories (names starting with `.`) when scanning.

#### Scenario: Save a new skill
- **WHEN** a skill entry is saved via `FileSkillStore.Save()`
- **THEN** the system SHALL create `<skillsDir>/<name>/SKILL.md` with YAML frontmatter and markdown body
- **AND** the skill directory SHALL be created with permission mode 0700

#### Scenario: Load active skills
- **WHEN** `FileSkillStore.ListActive()` is called
- **THEN** all skills with `status: active` in their frontmatter SHALL be returned
- **AND** directories whose name starts with `.` SHALL be skipped without logging a warning

#### Scenario: Hidden directory ignored
- **WHEN** `FileSkillStore.ListActive()` encounters a directory starting with `.`
- **THEN** it SHALL skip the directory silently without attempting to parse its contents

#### Scenario: Delete a skill
- **WHEN** `FileSkillStore.Delete()` is called with a skill name
- **THEN** the entire `<skillsDir>/<name>/` directory SHALL be removed

#### Scenario: SaveResource writes file to correct path
- **WHEN** `SaveResource` is called with skillName="my-skill" and relPath="scripts/run.sh"
- **THEN** the file SHALL be written to `<store-dir>/my-skill/scripts/run.sh`
- **AND** parent directories SHALL be created with permission mode 0700

### Requirement: SKILL.md Parsing
The system SHALL parse SKILL.md files with YAML frontmatter delimited by `---` lines, extracting metadata and body content.

#### Scenario: Parse valid SKILL.md
- **WHEN** a file with valid YAML frontmatter and markdown body is parsed
- **THEN** a `SkillEntry` SHALL be returned with all frontmatter fields populated and definition extracted from code blocks

#### Scenario: Parse file without frontmatter
- **WHEN** a file without `---` delimiters is parsed
- **THEN** an error SHALL be returned

### Requirement: Embedded Default Skills
The system SHALL embed default skill files via `//go:embed **/SKILL.md`. When no real skill SKILL.md files are present, a `.placeholder/SKILL.md` file SHALL exist to satisfy the embed glob pattern. The placeholder SHALL NOT contain valid YAML frontmatter and SHALL NOT be deployed as a usable skill. `EnsureDefaults()` SHALL skip any embedded path whose directory name starts with `.` (hidden directories).

#### Scenario: Build with no real default skills
- **WHEN** `go build` is run with only `.placeholder/SKILL.md` in the skills directory
- **THEN** the build SHALL succeed without errors

#### Scenario: Placeholder not deployed as skill
- **WHEN** `EnsureDefaults()` iterates over the embedded filesystem
- **THEN** entries whose directory name starts with `.` SHALL be skipped entirely
- **AND** no files from `.placeholder/` SHALL be written to the user's skills directory

#### Scenario: Future skill addition
- **WHEN** a new `skills/<name>/SKILL.md` file with valid frontmatter is added
- **THEN** it SHALL be automatically included in the embedded filesystem and deployed via `EnsureDefaults()`

#### Scenario: Existing skills preserved
- **WHEN** `EnsureDefaults()` is called and a skill directory already exists
- **THEN** that skill SHALL NOT be overwritten

#### Scenario: EnsureDefaults directory permissions
- **WHEN** `EnsureDefaults()` creates skill directories
- **THEN** all directories SHALL be created with permission mode 0700

### Requirement: Independent Skill Configuration
The system SHALL use a separate `SkillConfig` with `Enabled` and `SkillsDir` fields, independent of `KnowledgeConfig`.

#### Scenario: Skill system disabled
- **WHEN** `skill.enabled` is false in config
- **THEN** no skills SHALL be loaded and skill tools SHALL NOT be registered

#### Scenario: Custom skills directory
- **WHEN** `skill.skillsDir` is set to a custom path
- **THEN** skills SHALL be read from and written to that directory

### Requirement: SkillProvider Interface
The system SHALL decouple the `ContextRetriever` from skill storage via a `SkillProvider` interface.

#### Scenario: Skill provider wired
- **WHEN** a `SkillProvider` is set on the `ContextRetriever`
- **THEN** skill context items SHALL be retrieved via the provider instead of the knowledge store

#### Scenario: No skill provider
- **WHEN** no `SkillProvider` is configured
- **THEN** the skill layer SHALL return no items without error

### Requirement: Skill Registry
The system SHALL provide a registry for managing reusable skill definitions with lifecycle management.

#### Scenario: Create skill
- **WHEN** `CreateSkill` is called with a valid skill entry
- **THEN** the system SHALL validate the skill type is one of "composite", "script", "template", or "instruction"
- **AND** validate the skill definition matches the type requirements
- **AND** persist the skill with status "active"

#### Scenario: Valid instruction type
- **WHEN** a skill with type `instruction` is created
- **THEN** the registry SHALL accept it even with empty Definition

#### Scenario: Invalid skill type
- **WHEN** `CreateSkill` is called with an unrecognized skill type
- **THEN** the system SHALL return an error listing all valid types including `instruction`

#### Scenario: Composite skill validation
- **WHEN** creating a composite skill
- **THEN** the definition SHALL contain a "steps" array

#### Scenario: Script skill validation
- **WHEN** creating a script skill
- **THEN** the definition SHALL contain a "script" string
- **AND** the script SHALL be validated against dangerous patterns

#### Scenario: Template skill validation
- **WHEN** creating a template skill
- **THEN** the definition SHALL contain a "template" string

#### Scenario: Activate skill
- **WHEN** `ActivateSkill` is called with a skill name
- **THEN** the system SHALL set the skill status to "active"

#### Scenario: Load skills on startup
- **WHEN** the registry is initialized
- **THEN** `LoadSkills` SHALL load all active skills from the store

#### Scenario: App tool assembly with knowledge system
- **WHEN** the knowledge system is enabled and tools are assembled in `app.go`
- **THEN** the app SHALL use `LoadedSkills()` to append only dynamic skills
- **AND** SHALL NOT use `AllTools()` which would duplicate base tools already present in the tool list

### Requirement: Loaded Skills Retrieval
The registry SHALL provide a `LoadedSkills()` method that returns only dynamically loaded skill tools, excluding base tools.

#### Scenario: No skills loaded
- **WHEN** `LoadedSkills` is called before any skills are loaded
- **THEN** the system SHALL return an empty slice

#### Scenario: Skills loaded
- **WHEN** `LoadedSkills` is called after skills have been activated
- **THEN** the system SHALL return only the dynamically loaded skill tools
- **AND** the result SHALL NOT include any base tools passed during registry creation

#### Scenario: Concurrent safety
- **WHEN** `LoadedSkills` is called concurrently with `LoadSkills`
- **THEN** access SHALL be protected by a read lock

### Requirement: Skill Executor
The system SHALL safely execute skills of three types: composite, script, and template.

#### Scenario: Execute composite skill
- **WHEN** executing a composite skill
- **THEN** the system SHALL extract the steps array from the definition
- **AND** return an execution plan with step numbers, tool names, and parameters

#### Scenario: Execute script skill
- **WHEN** executing a script skill
- **THEN** the system SHALL validate the script against dangerous patterns
- **AND** create a temporary file via `os.CreateTemp` in the OS temp directory
- **AND** write the script content to the temp file and close it before execution
- **AND** execute it via `sh` with context-based timeout
- **AND** clean up the temporary file after execution via `defer os.Remove`

#### Scenario: Execute template skill
- **WHEN** executing a template skill
- **THEN** the system SHALL parse the template string as a Go text/template
- **AND** execute it with the provided parameters
- **AND** return the rendered output

### Requirement: Dangerous Pattern Validation
The system SHALL validate scripts against known dangerous patterns as a defense-in-depth measure.

#### Scenario: Reject dangerous scripts
- **WHEN** a script matches any dangerous pattern
- **THEN** the system SHALL return an error
- **AND** dangerous patterns SHALL include: recursive force delete (`rm -rf /`), fork bombs, pipe-to-shell (`curl|sh`), raw device writes (`>/dev/sd`), filesystem formatting (`mkfs.`), and raw disk copies (`dd if=`)

### Requirement: Skill Builder
The system SHALL provide a builder for constructing skill entries from tool execution traces.

#### Scenario: Build composite skill from steps
- **WHEN** `BuildFromSteps` is called with a name, description, and list of tool steps
- **THEN** the system SHALL construct a SkillEntry of type "composite" with the steps in the definition

#### Scenario: Build script skill
- **WHEN** `BuildScript` is called with a name, description, and script content
- **THEN** the system SHALL construct a SkillEntry of type "script" with the script in the definition

### Requirement: Executor Initialization
The system SHALL initialize the executor without filesystem side-effects.

#### Scenario: Infallible construction
- **WHEN** `NewExecutor` is called
- **THEN** the system SHALL return an `*Executor` value directly (no error)
- **AND** SHALL NOT create any directories or perform filesystem operations

### Requirement: Skill entry structure
`SkillEntry` SHALL include a `Source` string field to track the import origin URL and an `AllowedTools []string` field for pre-approved tool lists. The `Source` field SHALL be empty for locally created skills.

#### Scenario: Imported skill
- **WHEN** a skill is imported from an external URL
- **THEN** its `Source` field SHALL contain the origin URL

#### Scenario: Local skill
- **WHEN** a skill is created locally
- **THEN** its `Source` field SHALL be empty

#### Scenario: SkillEntry with AllowedTools
- **WHEN** a `SkillEntry` is created with `AllowedTools: ["exec", "fs_read"]`
- **THEN** the field SHALL be persisted through Save and restored through Get

### Requirement: Instruction skill type
The system SHALL support an `instruction` skill type for agent reference documents. Instruction skills store their entire markdown body as `definition["content"]`.

#### Scenario: Parse instruction SKILL.md
- **WHEN** a SKILL.md has no explicit `type` in frontmatter
- **THEN** the parser SHALL default to type `instruction` and store the body as content

#### Scenario: Parse instruction with explicit type
- **WHEN** a SKILL.md has `type: instruction` in frontmatter
- **THEN** the parser SHALL store the body as `definition["content"]`

### Requirement: Instruction skill as tool
Instruction skills SHALL be converted to agent tools. The tool description SHALL use the skill's original description (for agent reasoning). The handler SHALL return the full reference content, source, and description.

#### Scenario: Tool registration
- **WHEN** an instruction skill is loaded
- **THEN** it SHALL be registered as `skill_{name}` with the original description

#### Scenario: Tool invocation
- **WHEN** the agent calls an instruction skill tool
- **THEN** the handler SHALL return content, source, description, and type fields

#### Scenario: Empty description fallback
- **WHEN** an instruction skill has no description
- **THEN** the tool description SHALL default to "Reference guide for {name}"

### Requirement: Instruction skill execution
The executor SHALL handle the `instruction` type by returning the content directly without script execution.

#### Scenario: Execute instruction skill
- **WHEN** the executor is called with an instruction skill
- **THEN** it SHALL return a map with skill name, type, and content

### Requirement: Source field roundtrip in SKILL.md
The parser and renderer SHALL preserve the `source` field in YAML frontmatter through parse/render cycles.

#### Scenario: Source field roundtrip
- **WHEN** a SKILL.md with a `source` field is parsed and re-rendered
- **THEN** the re-parsed entry SHALL have the same `source` value

### Requirement: Instruction skill rendering
The renderer SHALL output instruction skill content as plain markdown (no code block wrapping).

#### Scenario: Render instruction
- **WHEN** an instruction skill is rendered to SKILL.md format
- **THEN** the body SHALL contain the raw content without code block delimiters

### Requirement: BuildInstructionSkill builder
The system SHALL provide a `BuildInstructionSkill` function that creates an instruction SkillEntry with name, description, content, and source.

#### Scenario: Build instruction skill
- **WHEN** `BuildInstructionSkill("name", "desc", "content", "https://source")` is called
- **THEN** it SHALL return a SkillEntry with type=instruction, status=active, createdBy=import

### Requirement: Registry Store accessor
The registry SHALL expose a `Store()` method returning the underlying `SkillStore` interface.

#### Scenario: Store access
- **WHEN** `registry.Store()` is called
- **THEN** it SHALL return the configured SkillStore instance
