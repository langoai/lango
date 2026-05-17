## MODIFIED Requirements

### Requirement: Cron job persistence
The system SHALL persist cron jobs in the Ent ORM with fields: id (UUID), name
(unique), schedule_type (at/every/cron), schedule, prompt, session_mode,
deliver_to ([]string), timezone, enabled, timeout, last_run_at, next_run_at, and
timestamps.

#### Scenario: CLI add updates an existing job by name
- **WHEN** `lango cron add` is run with the name of an existing cron job
- **THEN** the command SHALL update the existing job instead of creating a
  duplicate or returning a unique-constraint error

#### Scenario: Create a cron job with CLI timeout
- **WHEN** `lango cron add` is run with `--timeout 5m`
- **THEN** the persisted cron job SHALL store a five-minute per-job timeout

### Requirement: Isolated session execution
The system SHALL execute each cron job in an isolated agent session with a key
following the pattern "cron:<jobName>:<timestamp>" unless configuration or the
job explicitly selects the shared main session.

#### Scenario: CLI add uses configured default session mode
- **WHEN** `lango cron add` is run without an explicit `--isolated` flag
- **THEN** the created job SHALL use `cron.defaultSessionMode`
- **AND** `--isolated` help text SHALL describe the flag as an override rather
  than a static default

#### Scenario: CLI add falls back to isolated session mode
- **WHEN** `lango cron add` is run without an explicit `--isolated` flag
- **AND** `cron.defaultSessionMode` is empty
- **THEN** the created job SHALL use `session_mode="isolated"`

#### Scenario: CLI add can select shared main session mode
- **WHEN** `lango cron add` is run with `--isolated=false`
- **THEN** the created job SHALL use `session_mode="main"`

### Requirement: Cron CLI documented flag compatibility
The cron CLI SHALL accept the documented `--deliver-to`, `--timeout`, and
management `--id` flags while preserving existing positional and `--deliver`
forms.

#### Scenario: Add accepts deliver-to alias
- **WHEN** `lango cron add` is run with `--deliver-to telegram`
- **THEN** the created job SHALL use `telegram` as a delivery target

#### Scenario: Add rejects invalid timeout
- **WHEN** `lango cron add` is run with an invalid `--timeout` value
- **THEN** the command SHALL return an actionable parse error
- **AND** it SHALL NOT create a cron job

#### Scenario: Control commands accept id flag
- **WHEN** `lango cron pause`, `resume`, `delete`, or `history` is run with
  `--id <id-or-name>`
- **THEN** the command SHALL resolve the supplied selector as if it were passed
  positionally

#### Scenario: Control commands reject ambiguous selectors
- **WHEN** a cron control command receives both a positional selector and
  `--id`
- **THEN** the command SHALL return an actionable ambiguity error
