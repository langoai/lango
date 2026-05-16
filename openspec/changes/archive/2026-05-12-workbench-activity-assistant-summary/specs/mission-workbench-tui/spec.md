## ADDED Requirements

### Requirement: Workbench activity retains assistant reply summaries after a turn completes

The standalone workbench SHALL retain a concise assistant reply summary in its activity lane after a turn completes so the operator can still see what happened from the workbench shell.

#### Scenario: Successful turn adds assistant reply summary to activity
- **WHEN** a workbench turn completes successfully
- **THEN** the activity lane SHALL include a concise assistant reply summary for that turn

#### Scenario: Non-success turn adds outcome summary to activity
- **WHEN** a workbench turn completes with a non-success outcome
- **THEN** the activity lane SHALL include an outcome-labeled summary for that turn
