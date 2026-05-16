## ADDED Requirements

### Requirement: Public workbench docs mention Git-aware quick-start behavior

Public docs that describe standalone workbench starter prompts SHALL mention that the workbench can use live Git branch and dirty-state signals when available.

#### Scenario: Docs mention Git-aware starter prompts
- **WHEN** README or CLI/TUI docs describe context-aware starter prompts
- **THEN** they SHALL mention current-branch awareness
- **AND** they SHALL mention uncommitted-change awareness when Git metadata is available
