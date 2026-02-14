## ADDED Requirements

### Requirement: Blocked paths enforcement
The filesystem tool SHALL support a `BlockedPaths` configuration field. Any path that falls under a blocked path SHALL be denied with "access denied: protected path" before checking allowed paths.

#### Scenario: Access blocked path
- **WHEN** an agent attempts to read a file under `~/.lango/`
- **THEN** the system returns "access denied: protected path"

#### Scenario: Access path outside blocked paths
- **WHEN** an agent reads a file not under any blocked path
- **THEN** the file is read normally (subject to existing AllowedPaths checks)

#### Scenario: Empty blocked paths
- **WHEN** `BlockedPaths` is empty
- **THEN** no paths are blocked (existing behavior preserved)
