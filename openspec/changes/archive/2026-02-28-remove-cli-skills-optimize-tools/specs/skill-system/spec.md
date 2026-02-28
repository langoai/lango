## MODIFIED Requirements

### Requirement: Embedded Default Skills
The system SHALL embed default skill files via `//go:embed **/SKILL.md`. When no real skill SKILL.md files are present, a `.placeholder/SKILL.md` file SHALL exist to satisfy the embed glob pattern. The placeholder SHALL NOT contain valid YAML frontmatter and SHALL NOT be deployed as a usable skill.

#### Scenario: Build with no real default skills
- **WHEN** `go build` is run with only `.placeholder/SKILL.md` in the skills directory
- **THEN** the build SHALL succeed without errors

#### Scenario: Placeholder not deployed as skill
- **WHEN** `EnsureDefaults()` iterates over the embedded filesystem
- **THEN** the `.placeholder` directory SHALL be skipped or ignored because it lacks valid YAML frontmatter

#### Scenario: Future skill addition
- **WHEN** a new `skills/<name>/SKILL.md` file with valid frontmatter is added
- **THEN** it SHALL be automatically included in the embedded filesystem and deployed via `EnsureDefaults()`

## REMOVED Requirements

### Requirement: 42 default CLI wrapper skills
**Reason**: All 42 default skills wrapped `lango` CLI commands that require passphrase authentication, making them non-functional in agent mode. Built-in tools provide equivalent functionality.
**Migration**: Use built-in tools (exec, filesystem, crypto, secrets, cron, background, workflow, p2p, browser) instead. For CLI-only features (config, doctor, settings), run commands directly in the user's terminal.
