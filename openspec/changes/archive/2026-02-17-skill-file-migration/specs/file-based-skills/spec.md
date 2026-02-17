## ADDED Requirements

### Requirement: File-based skill storage
The system SHALL store skills as `<dir>/<name>/SKILL.md` files with YAML frontmatter containing name, description, type, status, and optional parameters.

#### Scenario: Save a new skill
- **WHEN** a skill entry is saved via `FileSkillStore.Save()`
- **THEN** the system creates `<skillsDir>/<name>/SKILL.md` with YAML frontmatter and markdown body

#### Scenario: Load an active skill
- **WHEN** `FileSkillStore.ListActive()` is called
- **THEN** all skills with `status: active` in their frontmatter are returned

#### Scenario: Delete a skill
- **WHEN** `FileSkillStore.Delete()` is called with a skill name
- **THEN** the entire `<skillsDir>/<name>/` directory is removed

### Requirement: SKILL.md parsing
The system SHALL parse SKILL.md files with YAML frontmatter delimited by `---` lines, extracting metadata and body content.

#### Scenario: Parse valid SKILL.md
- **WHEN** a file with valid YAML frontmatter and markdown body is parsed
- **THEN** a `SkillEntry` is returned with all frontmatter fields populated and definition extracted from code blocks

#### Scenario: Parse file without frontmatter
- **WHEN** a file without `---` delimiters is parsed
- **THEN** an error is returned

### Requirement: Embedded default skills
The system SHALL embed 30 default CLI skill files via `//go:embed` and deploy them to the user's skills directory on first run.

#### Scenario: First-run deployment
- **WHEN** `EnsureDefaults()` is called and the skills directory is empty
- **THEN** all 30 default skills are copied from the embedded filesystem to `<skillsDir>/<name>/SKILL.md`

#### Scenario: Existing skills preserved
- **WHEN** `EnsureDefaults()` is called and a skill directory already exists
- **THEN** that skill is NOT overwritten

### Requirement: Independent skill configuration
The system SHALL use a separate `SkillConfig` with `Enabled` and `SkillsDir` fields, independent of `KnowledgeConfig`.

#### Scenario: Skill system disabled
- **WHEN** `skill.enabled` is false in config
- **THEN** no skills are loaded and skill tools are not registered

#### Scenario: Custom skills directory
- **WHEN** `skill.skillsDir` is set to a custom path
- **THEN** skills are read from and written to that directory

### Requirement: SkillProvider interface for retriever
The system SHALL decouple the `ContextRetriever` from skill storage via a `SkillProvider` interface, allowing any implementation to supply skill metadata.

#### Scenario: Skill provider wired
- **WHEN** a `SkillProvider` is set on the `ContextRetriever`
- **THEN** skill context items are retrieved via the provider instead of the knowledge store

#### Scenario: No skill provider
- **WHEN** no `SkillProvider` is configured
- **THEN** the skill layer returns no items without error
