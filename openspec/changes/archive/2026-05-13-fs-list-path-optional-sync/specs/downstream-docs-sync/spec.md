## ADDED Requirements

### Requirement: Filesystem prompt guidance mentions fs_list default path

Prompt guidance for the filesystem tool SHALL mention that `fs_list` defaults to the current working directory when `path` is omitted.

#### Scenario: TOOL_USAGE mentions fs_list current-directory default
- **WHEN** an agent reads the filesystem section of `TOOL_USAGE.md`
- **THEN** it SHALL find that `fs_list` accepts an optional `path`
- **AND** that omitting `path` lists the current working directory
