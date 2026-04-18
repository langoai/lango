## ADDED Requirements

### Requirement: Learning history uses storage reader
The `lango learning history` command MUST read recent learning rows through a storage facade reader instead of querying Ent directly from the CLI layer.

#### Scenario: Learning history command reads through facade
- **WHEN** the user runs `lango learning history`
- **THEN** the command loads recent learning records from the storage facade reader
