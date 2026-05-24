## ADDED Requirements

### Requirement: Public workbench docs mention starter hotkeys

Public docs that describe the standalone workbench startup flow SHALL mention the ready-profile starter-prompt hotkeys once the product exposes them.

#### Scenario: Workbench docs mention starter hotkeys
- **WHEN** README or CLI/TUI docs describe the ready-profile workbench empty state
- **THEN** they SHALL mention that starter prompts are bound to `1`, `2`, and `3`
- **AND** they SHALL describe that those keys load prompts into the composer instead of leaving the prompts as passive copy
