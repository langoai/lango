## ADDED Requirements

### Requirement: Enter quick-start default follows repository state

The standalone workbench SHALL choose the default `Enter` quick-start prompt from workspace context instead of always using the summary prompt.

#### Scenario: Dirty repository defaults Enter to change review
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is a dirty Git repository
- **AND** the operator presses `Enter`
- **THEN** the seeded default prompt SHALL be the context-aware change-review prompt

#### Scenario: Clean workspace defaults Enter to summary
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the detected workspace is clean or not repository-backed
- **AND** the operator presses `Enter`
- **THEN** the seeded default prompt SHALL remain the summary-oriented default prompt
