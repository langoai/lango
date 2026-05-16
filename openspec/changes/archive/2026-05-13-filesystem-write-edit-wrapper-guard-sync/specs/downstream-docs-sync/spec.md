## ADDED Requirements

### Requirement: Filesystem prompt guidance mentions write/edit required inputs

Prompt guidance for the filesystem tool SHALL describe the required-input contract for `fs_write` and `fs_edit`.

#### Scenario: TOOL_USAGE mentions write/edit required inputs
- **WHEN** an agent reads the filesystem section of `TOOL_USAGE.md`
- **THEN** it SHALL find that `fs_write` requires `path` and `content`
- **AND** that `fs_edit` requires `path`, `startLine`, `endLine`, and `content`
