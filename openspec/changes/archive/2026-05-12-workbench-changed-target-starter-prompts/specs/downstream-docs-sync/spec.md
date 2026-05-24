## ADDED Requirements

### Requirement: Public workbench docs mention changed-target-aware quick-start behavior

Public docs that describe Git-aware workbench starter prompts SHALL mention changed-file or changed-directory awareness when that signal is available.

#### Scenario: Docs mention changed targets in starter prompts
- **WHEN** README or CLI/TUI docs describe dirty-repository starter prompt behavior
- **THEN** they SHALL mention changed-file or changed-directory awareness in addition to branch and dirty-state awareness
