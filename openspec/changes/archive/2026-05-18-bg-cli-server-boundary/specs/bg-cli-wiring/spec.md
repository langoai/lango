## MODIFIED Requirements

### Requirement: Background task CLI reports runtime boundary truthfully
The `lango bg` command family SHALL make the current runtime boundary explicit when no in-process background manager is available.

#### Scenario: Root CLI background command explains in-memory server boundary
- **WHEN** the root CLI is wired without an in-process background manager
- **AND** the user runs a `lango bg` subcommand
- **THEN** the command SHALL fail with an actionable message explaining that background task state is in-memory and owned by the running app/server process
- **AND** the message SHALL NOT imply that simply starting `lango serve` makes the current standalone CLI process able to inspect that manager

#### Scenario: In-process background command behavior remains available
- **WHEN** `internal/cli/bg.NewBgCmd` is constructed with a real background manager provider
- **THEN** `list`, `status`, `cancel`, and `result` SHALL continue to operate on that in-process manager
