## ADDED Requirements

### Requirement: Extension-owned skill subdirectories
The skill discovery walker SHALL recognize `ext-<pack-name>/` subdirectories under `skills.skillsDir` as the on-disk home for skills owned by an extension pack. Every skill under an `ext-<pack-name>/` subdir SHALL be discovered and registered in the same way as a non-ext skill at the same depth, preserving the existing SKILL.md parsing and frontmatter conventions.

#### Scenario: Pack-owned skill discovered
- **WHEN** `<skillsDir>/ext-python-dev/pytest-refactor/SKILL.md` exists with valid frontmatter
- **THEN** `FileSkillStore.ListActive()` SHALL include a `SkillEntry` for `pytest-refactor` derived from that file

#### Scenario: Existing hidden-dir rule still applies
- **WHEN** `<skillsDir>/ext-python-dev/.cache/` exists as a hidden directory
- **THEN** the walker SHALL skip the hidden directory silently, unchanged from existing behavior

### Requirement: User-authored skill name takes precedence over extension-authored
When a skill name is provided by both a user-authored directory (directly under `skillsDir`) and an extension-authored directory (under `<skillsDir>/ext-<pack>/`), the skill registry SHALL prefer the user-authored entry. The extension-authored entry SHALL remain on disk but SHALL NOT be returned from lookup by that name. The registry SHALL log a debug-level `skill.name.shadowed_by_user` message naming the pack.

#### Scenario: User override wins
- **WHEN** `<skillsDir>/pytest-refactor/` and `<skillsDir>/ext-python-dev/pytest-refactor/` both exist
- **THEN** a lookup for `pytest-refactor` SHALL return the user-authored entry
- **AND** a debug log SHALL be emitted naming `python-dev` as the shadowed source

### Requirement: Cross-extension skill collision is not resolvable at runtime
When two extension packs have each written a skill with the same name into their respective `ext-<pack>/` subdirs, the skill registry SHALL return an error at load time naming both packs and the colliding name. The install contract (see `extension-pack-core`) SHALL prevent this state from occurring through fresh installs, but this runtime guard catches the state on an upgrade, manual edit, or from-prior-version filesystem.

#### Scenario: Collision raises at load
- **WHEN** `<skillsDir>/ext-python-A/foo/` and `<skillsDir>/ext-python-B/foo/` both exist
- **THEN** registry construction SHALL return an error naming both pack prefixes and the skill name `foo`
- **AND** the caller (startup wiring) SHALL surface this as a fatal error so the user must resolve it before the app proceeds

### Requirement: Skill source attribution (additive field)
`SkillEntry` SHALL gain an optional `SourcePack string` field carrying the pack name for extension-authored skills. For user-authored or built-in skills, `SourcePack` SHALL be empty. The field SHALL marshal with `omitempty` in JSON and SHALL NOT alter the existing SKILL.md on-disk format.

#### Scenario: Extension skill carries source
- **WHEN** a skill is loaded from `<skillsDir>/ext-python-dev/pytest-refactor/SKILL.md`
- **THEN** the returned `SkillEntry.SourcePack` SHALL equal `python-dev`

#### Scenario: User skill has empty source
- **WHEN** a skill is loaded from `<skillsDir>/my-skill/SKILL.md`
- **THEN** the returned `SkillEntry.SourcePack` SHALL be the empty string
