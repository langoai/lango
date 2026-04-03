## MODIFIED Requirements

### Requirement: Delete operation handles symlinks safely
The `Delete` operation SHALL use a symlink-aware validation flow. When the target path is a symlink, Delete SHALL validate the link's location (not the resolved target) against blocked/allowed directories, then remove the symlink itself. When the target is not a symlink, Delete SHALL use the standard `validatePath` flow.

#### Scenario: Delete symlink removes link not target
- **WHEN** `Delete` is called on a path that is a symlink
- **THEN** the symlink itself SHALL be removed and the target file SHALL remain intact

#### Scenario: Delete symlink in blocked directory
- **WHEN** `Delete` is called on a symlink located in a blocked directory
- **THEN** the operation SHALL be denied regardless of where the symlink target points

#### Scenario: Delete symlink pointing to blocked target
- **WHEN** `Delete` is called on a symlink in an allowed directory that points to a blocked target
- **THEN** the symlink SHALL be deleted (since only the link is removed, not the target)

#### Scenario: OS alias canonicalization for symlink location
- **WHEN** the symlink's parent directory involves an OS alias (e.g., macOS `/var` → `/private/var`)
- **THEN** the parent directory SHALL be resolved via `EvalSymlinks` before blocked/allowed comparison

### Requirement: Path access check compares both resolved and unresolved config entries
The `checkPathAccess` function SHALL compare the input path against both the unresolved and resolved versions of each `BlockedPaths` and `AllowedPaths` config entry. This handles cases where the config entry itself is a symlink.

#### Scenario: Config entry is a symlink
- **WHEN** `BlockedPaths` contains a path that is itself a symlink
- **THEN** the block check SHALL match against both the symlink path and its resolved target
