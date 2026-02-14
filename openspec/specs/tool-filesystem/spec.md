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
