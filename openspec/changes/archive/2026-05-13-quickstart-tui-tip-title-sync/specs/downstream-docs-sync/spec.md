## ADDED Requirements

### Requirement: Quickstart TUI tip title does not imply a chat-only surface
The public quickstart guide SHALL not title the bare-`lango` entrypoint tip as if it were a chat-only surface.

#### Scenario: Quickstart tip title uses neutral TUI wording
- **WHEN** a user reads the interactive TUI tip in `docs/getting-started/quickstart.md`
- **THEN** the tip title SHALL use neutral wording such as `Interactive TUI`
- **AND** it SHALL NOT title the workbench entrypoint as `Interactive TUI Chat`
