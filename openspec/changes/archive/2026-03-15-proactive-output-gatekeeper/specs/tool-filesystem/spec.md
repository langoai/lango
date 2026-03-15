## ADDED Requirements

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
