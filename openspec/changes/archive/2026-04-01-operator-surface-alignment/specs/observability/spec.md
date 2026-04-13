## ADDED Requirements

### Requirement: Policy metrics gateway endpoint
The system SHALL expose a `/metrics/policy` HTTP endpoint that returns policy decision statistics including total block count, total observe count, and per-reason breakdown counts as JSON.

#### Scenario: Fetch policy metrics via gateway
- **WHEN** a GET request is made to `/metrics/policy` on the gateway
- **THEN** the response SHALL be a JSON object with `blocks` (int), `observes` (int), and `byReason` (map of reason string to count) fields

### Requirement: Policy metrics CLI command
The system SHALL provide a `lango metrics policy` CLI command that displays policy decision statistics from the gateway endpoint. The command SHALL support `--output table|json` and `--addr` flags consistent with other metrics subcommands.

#### Scenario: Display policy metrics in table format
- **WHEN** the user runs `lango metrics policy`
- **THEN** the CLI SHALL display block count, observe count, and a per-reason breakdown table

#### Scenario: Display policy metrics in JSON format
- **WHEN** the user runs `lango metrics policy --output json`
- **THEN** the CLI SHALL output the raw JSON from the `/metrics/policy` endpoint

### Requirement: Policy decision audit logging
The system SHALL record policy decision events (observe and block verdicts) to the audit log database. Each audit entry SHALL include session key, verdict, reason, the original command, and the unwrapped command.

#### Scenario: Block verdict written to audit log
- **WHEN** the exec policy evaluator emits a `PolicyDecisionEvent` with verdict "block"
- **THEN** the audit recorder SHALL write an entry with action `policy_decision`, the actor (agent name or "system"), and details containing verdict, reason, unwrapped command, and message

### Requirement: Recovery decision event observability
The system SHALL publish a `RecoveryDecisionEvent` on the event bus when a recovery decision is made. The event SHALL include cause class, action taken, attempt number, backoff duration, and session key.

#### Scenario: Recovery decision event emitted on retry
- **WHEN** the coordinating executor decides to retry after an agent failure
- **THEN** a `RecoveryDecisionEvent` SHALL be published with the classified cause class, action "retry" or "retry_with_hint", the current attempt number, and the computed backoff duration

### Requirement: Recovery exponential backoff documentation
The observability documentation SHALL describe the exponential backoff formula (min(baseDelay * 2^attempt, maxBackoff)) with base delay of 1 second and maximum delay of 30 seconds.

#### Scenario: Operator reads backoff documentation
- **WHEN** the operator consults the observability docs recovery section
- **THEN** the documentation SHALL state the backoff formula, base delay, max delay, and the per-error-class retry limits table

### Requirement: Per-error-class retry limits documentation
The observability documentation SHALL describe the per-error-class retry limit system, listing the default limits for rate_limit (5), transient (3), malformed_tool_call (1), and timeout (3) cause classes.

#### Scenario: Operator reads retry limit documentation
- **WHEN** the operator consults the observability docs recovery section
- **THEN** the documentation SHALL include a table of cause classes and their default maximum retry counts
