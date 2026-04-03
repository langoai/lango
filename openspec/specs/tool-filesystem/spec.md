## ADDED Requirements

### Requirement: File reading
The system SHALL read file contents with encoding detection and size limits.

#### Scenario: Read text file
- **WHEN** reading a text file
- **THEN** the content SHALL be returned as a string with detected encoding

#### Scenario: Read binary file
- **WHEN** reading a binary file
- **THEN** the content SHALL be returned as base64 or indicate binary nature

#### Scenario: File size limit
- **WHEN** a file exceeds the configured size limit
- **THEN** an error SHALL be returned with the actual size

### Requirement: File writing
The system SHALL write content to files with atomic write support.

#### Scenario: Write new file
- **WHEN** writing to a non-existent path
- **THEN** the file SHALL be created with specified content

#### Scenario: Overwrite existing file
- **WHEN** writing to an existing file
- **THEN** the content SHALL replace the previous content

#### Scenario: Create parent directories
- **WHEN** parent directories do not exist
- **THEN** they SHALL be created automatically

### Requirement: File editing
The system SHALL support surgical edits to existing files.

#### Scenario: Line range replacement
- **WHEN** editing with a line range and replacement content
- **THEN** only the specified lines SHALL be replaced

#### Scenario: Search and replace
- **WHEN** editing with a search pattern and replacement
- **THEN** matching content SHALL be replaced

### Requirement: Directory operations
The system SHALL support listing and navigating directories.

#### Scenario: List directory contents
- **WHEN** listing a directory
- **THEN** files and subdirectories SHALL be returned with metadata

#### Scenario: Recursive listing
- **WHEN** listing with recursive option
- **THEN** all nested contents SHALL be included up to depth limit

#### Scenario: Delete file or directory
- **WHEN** deletion is requested for a path
- **THEN** the system SHALL remove the target and its contents if it is a directory

### Requirement: Path safety
The system SHALL validate file paths to prevent directory traversal attacks.

#### Scenario: Path traversal attempt
- **WHEN** a path contains ".." to escape allowed directory
- **THEN** the operation SHALL be rejected with an error

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

### Requirement: File metadata inspection
The system SHALL provide a `fs_stat` tool that returns file metadata (path, size, line count, modification time, isDir, permission) without reading the file content.

#### Scenario: Stat a regular file
- **WHEN** `fs_stat` is called with a path to a regular file
- **THEN** the result SHALL include size, line count, modTime, and permission

#### Scenario: Stat a directory
- **WHEN** `fs_stat` is called with a path to a directory
- **THEN** `isDir` SHALL be true and `lines` SHALL be 0

### Requirement: Partial file reading
The system SHALL support optional `offset` (1-indexed line number) and `limit` (max lines) parameters on `fs_read`. When provided, the result SHALL include `totalLines` and `size` metadata.

#### Scenario: Read with offset and limit
- **WHEN** `fs_read` is called with `offset=3` and `limit=2` on a 5-line file
- **THEN** lines 3-4 SHALL be returned with `totalLines=5`

#### Scenario: Read without offset/limit (backward compatible)
- **WHEN** `fs_read` is called without offset or limit parameters
- **THEN** the full file content SHALL be returned as a plain string (same as before)


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
The `Delete` method SHALL accept `context.Context` as its first parameter. In P2P context (`ctxkeys.IsP2PRequest(ctx)`), deletion MUST use `os.Remove` (single file or empty directory only) instead of `os.RemoveAll` (recursive). The filesystem package MUST NOT define its own P2P context key.

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

### Requirement: Delete operation handles symlinks safely
The `Delete` operation SHALL use a symlink-aware validation flow. When the target path is a symlink (detected via `os.Lstat` before `validatePath`), Delete SHALL resolve only the parent directory via `EvalSymlinks`, validate the canonical link location against blocked/allowed directories via `checkPathAccess`, then remove the symlink itself. When the target is not a symlink, Delete SHALL use the standard `validatePath` flow.

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

## REMOVED Requirements

### Requirement: P2P context detection
**Reason**: Replaced by canonical `ctxkeys.WithP2PRequest`/`ctxkeys.IsP2PRequest` from the `ctxkeys` package.
**Migration**: Use `ctxkeys.IsP2PRequest(ctx)` instead of `filesystem.IsP2PContext(ctx)`.
