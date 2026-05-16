## ADDED Requirements

### Requirement: Public workbench docs mention post-turn next-step defaults

Public workbench documentation SHALL explain that the empty workbench changes its default `Enter` starter after a completed turn.

#### Scenario: Docs mention post-turn next-step starter
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that, after a turn completes, the empty workbench defaults `Enter` to the next-step starter instead of returning to the original summary starter
