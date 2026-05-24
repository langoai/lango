## MODIFIED Requirements

### Requirement: Cross-extension skill collision is not resolvable at runtime
When two extension packs have each written a skill with the same name into their respective `ext-<pack>/` subdirs, the skill registry SHALL return an error at load time naming both packs and the colliding name. The install contract (see `extension-pack-core`) SHALL prevent this state from occurring through fresh installs, but this runtime guard catches the state on an upgrade, manual edit, or from-prior-version filesystem.

#### Scenario: Collision raises at load
- **WHEN** `<skillsDir>/ext-python-A/foo/` and `<skillsDir>/ext-python-B/foo/` both exist
- **THEN** registry construction SHALL return an error naming both pack prefixes and the skill name `foo`
- **AND** the caller (startup wiring) SHALL surface this as a fatal error so the user must resolve it before the app proceeds

#### Scenario: App startup fails on extension skill collision
- **WHEN** the app intelligence module loads an extension registry with two healthy packs that both provide active skill `foo`
- **THEN** startup SHALL return an error
- **AND** the error SHALL include the colliding skill name
- **AND** the error SHALL include both source pack names
