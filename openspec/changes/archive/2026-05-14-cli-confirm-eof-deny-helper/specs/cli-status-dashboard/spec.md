## ADDED Requirements

### Requirement: Dead-letter retry EOF uses shared deny behavior
`lango status dead-letter retry` SHALL use the shared EOF-deny confirmation helper for its default confirmation path.

#### Scenario: Dead-letter retry EOF aborts cleanly through shared helper
- **WHEN** the retry confirmation reaches EOF before approval
- **THEN** the command SHALL abort the retry without invoking the mutation path
