## ADDED Requirements

### Requirement: README top-level command summaries stay stream-aware
README.md SHALL keep its top-level command summaries aligned with the current stdout/stderr routing contracts when those contracts are already part of the tested CLI surface.

#### Scenario: README mentions top-level stream contracts
- **WHEN** top-level utility and TUI entrypoint stream-routing behavior is part of the current tested contract
- **THEN** README SHALL summarize the stdout/stderr routing at a high level
