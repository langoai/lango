## MODIFIED Requirements

### Requirement: DenyPaths must be existing directories
The `compileBwrapArgs` function SHALL stat each `DenyPath` and return a clear error when the path does not exist or is neither a directory nor a regular file. Directory deny paths SHALL be added to the argv as `--tmpfs <path>`. Regular file deny paths SHALL be added to the argv as `--ro-bind /dev/null <path>` — read operations on the file yield EOF, write operations return EACCES, and the parent directory structure is preserved so the file still appears to exist. Device nodes, sockets, and fifos SHALL produce an error with the message `"unsupported file mode"`.

(Note: the requirement name "DenyPaths must be existing directories" is preserved for backward compatibility with archived delta specs; despite the name, regular files are now also supported via the `/dev/null` bind trick. PR 5c relaxed the pre-existing "directory only" restriction.)

#### Scenario: Directory deny path is masked with tmpfs
- **WHEN** `policy.Filesystem.DenyPaths` contains an existing directory
- **THEN** the returned slice SHALL contain `--tmpfs <path>`

#### Scenario: File deny path is rejected
- **WHEN** `policy.Filesystem.DenyPaths` contains the path of an existing regular file
- **THEN** the returned slice SHALL contain `--ro-bind /dev/null <path>` (PR 5c: file-level deny now supported; pre-5c this scenario rejected the file with "must be a directory")
- **AND** the returned slice SHALL NOT contain `--tmpfs <path>` for that entry

#### Scenario: Missing deny path is rejected
- **WHEN** `policy.Filesystem.DenyPaths` contains a path that does not exist
- **THEN** `compileBwrapArgs` SHALL return an error referencing the missing path

#### Scenario: Unsupported file mode is rejected
- **WHEN** `policy.Filesystem.DenyPaths` contains a device node, socket, or fifo
- **THEN** `compileBwrapArgs` SHALL return an error containing `"unsupported file mode"`

## ADDED Requirements

### Requirement: Path entries flow through shared normalizePath pipeline
All `compileBwrapArgs` path classes (`ReadPaths`, `WritePaths`, `DenyPaths`) SHALL route each entry through the shared `normalizePath` helper in `internal/sandbox/os/policy.go` before emitting bwrap flags. The helper implements the canonical pipeline `sanitize → filepath.Abs → filepath.Glob → filepath.EvalSymlinks (with nonexistent fallback)` and returns zero or more concrete filesystem paths. Each returned path is then processed by the path-class-specific emission logic (`--ro-bind` for reads, `--bind` for writes, `--tmpfs`/`--ro-bind /dev/null` for denies).

This guarantees that all three path classes support:
- **Glob patterns** (`*`, `?`, `[`): expanded via `filepath.Glob`. Zero matches silently skip; invalid patterns return `filepath.ErrBadPattern` wrapped in a sandbox error.
- **Symlinks**: resolved via `filepath.EvalSymlinks`. Nonexistent paths fall back to the pre-resolve absolute path so downstream `os.Stat` catches the missing-path error with the existing error message format.
- **Injection character rejection**: preserved from the existing `sanitizePath` step — double-quote, parenthesis, semicolon, newline continue to return `ErrInvalidPolicy`.

#### Scenario: Glob pattern in DenyPaths expands to matches
- **WHEN** `policy.Filesystem.DenyPaths` contains an entry like `/tmp/fixture/*.log` AND the glob matches three files
- **THEN** the returned slice SHALL contain three `--ro-bind /dev/null` entries, one per matched file
- **AND** the original glob pattern SHALL NOT appear in the returned slice

#### Scenario: Symlinked DenyPath resolves to target
- **WHEN** `policy.Filesystem.DenyPaths` contains a symlink pointing to an existing directory
- **THEN** the returned slice SHALL contain `--tmpfs <resolved-target>`, NOT `--tmpfs <symlink-path>`

#### Scenario: Unmatched glob is silently skipped
- **WHEN** `policy.Filesystem.DenyPaths` contains a glob pattern with zero matches
- **THEN** the returned slice SHALL NOT contain any entry derived from that pattern
- **AND** `compileBwrapArgs` SHALL NOT return an error for that entry

#### Scenario: Invalid glob pattern is rejected at startup
- **WHEN** `policy.Filesystem.DenyPaths` contains a syntactically invalid glob pattern like `/tmp/[unclosed`
- **THEN** `compileBwrapArgs` SHALL return an error containing `"invalid glob pattern"`
