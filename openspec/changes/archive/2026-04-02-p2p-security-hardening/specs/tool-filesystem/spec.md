## MODIFIED Requirements

### Requirement: Path safety
The system SHALL validate file paths using `filepath.EvalSymlinks()` after `filepath.Abs()` to resolve symlinks before checking against allowed and blocked path lists. Both the target path and the config paths (allowed/blocked) MUST be resolved through `EvalSymlinks` to handle OS-specific symlink directories (e.g., macOS `/var` → `/private/var`).

#### Scenario: Symlink escape blocked
- **GIVEN** an allowed path `/workspace`
- **WHEN** a file at `/workspace/link` symlinks to `/etc/passwd`
- **THEN** the resolved path `/etc/passwd` is checked against allowed paths
- **AND** access is denied because `/etc/passwd` is outside `/workspace`

#### Scenario: Symlink within allowed directory
- **GIVEN** an allowed path `/workspace`
- **WHEN** a file at `/workspace/link` symlinks to `/workspace/data/file.txt`
- **THEN** access is allowed because the resolved path is within `/workspace`

#### Scenario: Broken symlink handled gracefully
- **WHEN** `filepath.EvalSymlinks()` fails (target does not exist)
- **THEN** validation continues with the cleaned absolute path (no error)

### Requirement: Directory operations
The `Delete` method SHALL accept `context.Context` as its first parameter. In P2P context (`IsP2PContext(ctx)`), deletion MUST use `os.Remove` (single file or empty directory only) instead of `os.RemoveAll` (recursive).

#### Scenario: P2P delete single file
- **WHEN** deletion is requested from a P2P context for a regular file
- **THEN** the file is deleted via `os.Remove`

#### Scenario: P2P delete non-empty directory blocked
- **WHEN** deletion is requested from a P2P context for a non-empty directory
- **THEN** `os.Remove` fails with "directory not empty"
- **AND** recursive deletion does NOT occur

#### Scenario: Local delete unchanged
- **WHEN** deletion is requested from a local (non-P2P) context
- **THEN** `os.RemoveAll` is used as before (backward compatible)

## ADDED Requirements

### Requirement: P2P context detection
The `filesystem` package MUST provide `WithP2PContext(ctx)` and `IsP2PContext(ctx)` functions for P2P origin marking at the package level.
