## ADDED Requirements

### Requirement: Directory listing keeps path optional

The `fs_list` tool SHALL accept an optional `path` parameter and SHALL default to the current working directory when `path` is omitted.

#### Scenario: fs_list defaults to the current directory
- **WHEN** `fs_list` is invoked without `path`
- **THEN** the tool SHALL list the current working directory successfully
