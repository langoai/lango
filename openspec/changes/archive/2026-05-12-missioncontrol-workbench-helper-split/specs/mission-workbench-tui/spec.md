## ADDED Requirements

### Requirement: Mission Control isolates workbench-only quick-start helpers from generic page flow

The workbench-specific quick-start and setup-recovery helper layer SHALL remain isolated from the generic Mission Control page flow so future workbench UX changes do not further crowd the shared page source.

#### Scenario: Workbench helpers live in a dedicated companion source
- **WHEN** the Mission Control page source is organized for maintenance
- **THEN** workbench-only starter/setup helper logic SHALL live in a dedicated companion source file rather than continuing to accumulate inside the primary Mission Control page source
