## ADDED Requirements

### Requirement: Skill directory hash coverage
The system SHALL hash ALL files in a skill directory (not just the manifest-listed file) when `contents.skills[].path` points at a `SKILL.md` file or a directory. The `FileHashes` in `.installed` metadata MUST cover every file that `copySkillsToStore` and `copyPackFiles` copy.

#### Scenario: SKILL.md with sibling files
- **WHEN** a manifest declares `path: skills/foo/SKILL.md` and `skills/foo/` contains `SKILL.md` plus `references/guide.md`
- **THEN** `fetchFromDir` SHALL hash both `skills/foo/SKILL.md` and `skills/foo/references/guide.md`
- **AND** the `.installed` metadata SHALL contain hashes for both files

#### Scenario: Directory-type skill path
- **WHEN** a manifest declares `path: skills/foo/` (directory, not a file)
- **THEN** `fetchFromDir` SHALL walk the directory and hash all regular files
- **AND** `hashFile` SHALL NOT be called with a directory path (no `os.ReadFile` on directories)

### Requirement: Pack mirror copies full skill directories
The system SHALL copy the full skill directory (not just the listed file) into the pack mirror during `Install`. The pack mirror content MUST match the hash coverage exactly.

#### Scenario: Install pack with skill directory
- **WHEN** installing a pack whose manifest lists `path: skills/foo/SKILL.md`
- **THEN** `copyPackFiles` SHALL copy the entire `skills/foo/` directory tree into the staging dir
- **AND** on next startup, tamper detection SHALL NOT report false positives for a cleanly installed pack

### Requirement: AllowedExtPacks integrity enforcement
The skill walker (`FileSkillStore.ListActive`) SHALL only walk `ext-<pack>/` subdirectories whose pack name appears in the `AllowedExtPacks` set. When `AllowedExtPacks` is nil, ALL ext-packs SHALL be skipped.

#### Scenario: Extensions disabled
- **WHEN** `extensions.enabled=false` (no extension registry loaded)
- **THEN** `AllowedExtPacks` SHALL be nil
- **AND** `ListActive` SHALL skip all `ext-*` subdirectories

#### Scenario: Tampered pack excluded
- **WHEN** `extensions.enforceIntegrity=true` and pack "foo" fails tamper detection
- **THEN** "foo" SHALL NOT appear in `AllowedExtPacks`
- **AND** skills under `ext-foo/` SHALL NOT be loaded

### Requirement: Extension prompt source wiring
The system SHALL read `Registry.PromptSources()` at startup and inject each prompt file as a `StaticSection` (priority 850) into the prompt builder.

#### Scenario: Pack with prompts section
- **WHEN** a healthy pack declares `contents.prompts` with a valid file
- **THEN** the file content SHALL appear in the runtime system prompt as a section with ID `extension_<pack>_<section>`

## MODIFIED Requirements

### Requirement: view_skill resolves extension-owned paths
The `view_skill` tool SHALL resolve skill file paths using the `SourcePack` field from the skill entry. Extension-owned skills at `<skillsDir>/ext-<pack>/<name>/` SHALL be correctly located.

#### Scenario: View extension-owned skill
- **WHEN** `view_skill` is called with a skill name that has `SourcePack="mypack"`
- **THEN** the skill root SHALL be resolved as `<skillsDir>/ext-mypack/<name>/`
- **AND** the file content SHALL be returned successfully
