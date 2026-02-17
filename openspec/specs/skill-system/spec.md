## ADDED Requirements

### Requirement: File-Based Skill Storage
The system SHALL store skills as `<dir>/<name>/SKILL.md` files with YAML frontmatter containing name, description, type, status, and optional parameters.

#### Scenario: Save a new skill
- **WHEN** a skill entry is saved via `FileSkillStore.Save()`
- **THEN** the system SHALL create `<skillsDir>/<name>/SKILL.md` with YAML frontmatter and markdown body

#### Scenario: Load active skills
- **WHEN** `FileSkillStore.ListActive()` is called
- **THEN** all skills with `status: active` in their frontmatter SHALL be returned

#### Scenario: Delete a skill
- **WHEN** `FileSkillStore.Delete()` is called with a skill name
- **THEN** the entire `<skillsDir>/<name>/` directory SHALL be removed

### Requirement: SKILL.md Parsing
The system SHALL parse SKILL.md files with YAML frontmatter delimited by `---` lines, extracting metadata and body content.

#### Scenario: Parse valid SKILL.md
- **WHEN** a file with valid YAML frontmatter and markdown body is parsed
- **THEN** a `SkillEntry` SHALL be returned with all frontmatter fields populated and definition extracted from code blocks

#### Scenario: Parse file without frontmatter
- **WHEN** a file without `---` delimiters is parsed
- **THEN** an error SHALL be returned

### Requirement: Embedded Default Skills
The system SHALL embed 30 default CLI skill files via `//go:embed` and deploy them to the user's skills directory on first run.

#### Scenario: First-run deployment
- **WHEN** `EnsureDefaults()` is called and a skill directory does not exist
- **THEN** the default skill SHALL be copied from the embedded filesystem to `<skillsDir>/<name>/SKILL.md`

#### Scenario: Existing skills preserved
- **WHEN** `EnsureDefaults()` is called and a skill directory already exists
- **THEN** that skill SHALL NOT be overwritten

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
- **THEN** the system SHALL validate the skill type is one of "composite", "script", or "template"
- **AND** validate the skill definition matches the type requirements
- **AND** persist the skill with status "active"

#### Scenario: Invalid skill type
- **WHEN** `CreateSkill` is called with an unrecognized skill type
- **THEN** the system SHALL return an error

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
