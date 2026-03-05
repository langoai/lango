## MODIFIED Requirements

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

#### Scenario: EnsureDefaults directory permissions
- **WHEN** `EnsureDefaults()` creates skill directories
- **THEN** all directories SHALL be created with permission mode 0700
