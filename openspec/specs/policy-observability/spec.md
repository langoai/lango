## Purpose

Capability spec for policy-observability. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Audit recorder subscribes to PolicyDecisionEvent
The audit recorder SHALL subscribe to `PolicyDecisionEvent` via `SubscribeTyped` and persist each event as an AuditLog entry with action `policy_decision`, actor set to AgentName, target set to Command, and details containing verdict, reason, and unwrapped command.

#### Scenario: Block verdict creates audit entry
- **WHEN** a PolicyDecisionEvent with verdict="block" is published
- **THEN** the audit recorder SHALL create an AuditLog entry with action=policy_decision, actor=AgentName, target=Command, and details containing verdict="block", reason, and unwrapped command

#### Scenario: Observe verdict creates audit entry
- **WHEN** a PolicyDecisionEvent with verdict="observe" is published
- **THEN** the audit recorder SHALL create an AuditLog entry with action=policy_decision, actor=AgentName, target=Command, and details containing verdict="observe", reason, and unwrapped command

### Requirement: Turntrace defines EventPolicyDecision constant
The turntrace package SHALL define `EventPolicyDecision EventType = "policy_decision"` for use in trace timeline integration.

#### Scenario: Constant exists with correct value
- **WHEN** the turntrace package is imported
- **THEN** `EventPolicyDecision` SHALL equal `"policy_decision"`

### Requirement: MetricsCollector records policy decisions
The MetricsCollector SHALL provide a `RecordPolicyDecision(verdict, reason string)` method that increments block or observe counters and aggregates counts by reason code.

#### Scenario: Block verdict increments block counter
- **WHEN** `RecordPolicyDecision("block", "lango_cli")` is called
- **THEN** the block counter SHALL increment by 1
- **AND** the byReason map SHALL increment "lango_cli" by 1

#### Scenario: Observe verdict increments observe counter
- **WHEN** `RecordPolicyDecision("observe", "opaque_pattern")` is called
- **THEN** the observe counter SHALL increment by 1
- **AND** the byReason map SHALL increment "opaque_pattern" by 1

#### Scenario: Thread safety under concurrent access
- **WHEN** multiple goroutines call RecordPolicyDecision concurrently
- **THEN** all counters SHALL be accurate with no data races

### Requirement: SystemSnapshot includes PolicyMetrics
The SystemSnapshot struct SHALL include a PolicyMetrics field containing Blocks count, Observes count, and ByReason map.

#### Scenario: Snapshot reflects recorded decisions
- **WHEN** RecordPolicyDecision has been called with various verdicts and reasons
- **THEN** Snapshot().Policy SHALL return a PolicyMetrics with accurate Blocks, Observes, and ByReason values

### Requirement: HTTP endpoint exposes policy metrics
The observability routes SHALL include a GET `/metrics/policy` endpoint that returns PolicyMetrics as JSON from the collector snapshot.

#### Scenario: Policy metrics returned as JSON
- **WHEN** a GET request is made to `/metrics/policy`
- **THEN** the response SHALL be JSON containing blocks, observes, and byReason fields

### Requirement: CLI subcommand for policy metrics
The CLI SHALL provide a `lango metrics policy` subcommand that fetches from `/metrics/policy` and renders as table (default) or JSON.

#### Scenario: Table output
- **WHEN** `lango metrics policy` is run without --output flag
- **THEN** it SHALL display blocks and observes counts plus a reason breakdown table

#### Scenario: JSON output
- **WHEN** `lango metrics policy --output json` is run
- **THEN** it SHALL output raw JSON from the endpoint

### Requirement: Wiring subscribes to PolicyDecisionEvent
The `initObservability` function SHALL subscribe to PolicyDecisionEvent on the event bus and call `collector.RecordPolicyDecision` for each event.

#### Scenario: Subscription wired when observability enabled
- **WHEN** observability is enabled in config
- **THEN** PolicyDecisionEvent subscriptions SHALL be registered on the event bus
- **AND** each event SHALL be forwarded to the metrics collector
