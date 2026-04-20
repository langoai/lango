## Purpose

Capability spec for exportability-policy. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Source-primary exportability evaluation
The system SHALL evaluate artifact exportability from source lineage rather than final content alone. The evaluator SHALL support the source classes `public`, `user-exportable`, and `private-confidential`, and SHALL emit one of `exportable`, `blocked`, or `needs-human-review`.

#### Scenario: Public-only artifact is exportable
- **WHEN** an artifact is evaluated from only `public` sources
- **THEN** the evaluator SHALL return `exportable`

#### Scenario: User-exportable artifact is exportable
- **WHEN** an artifact is evaluated from one or more `user-exportable` sources and no `private-confidential` source
- **THEN** the evaluator SHALL return `exportable`

#### Scenario: Private source blocks artifact
- **WHEN** any contributing source is `private-confidential`
- **THEN** the evaluator SHALL return `blocked`

#### Scenario: Missing source metadata requires human review
- **WHEN** any contributing source lacks resolvable source classification metadata
- **THEN** the evaluator SHALL return `needs-human-review`

### Requirement: Exportability decision receipts
Each exportability evaluation SHALL produce a receipt-style decision record containing the decision stage, decision state, policy code, human-readable explanation, and source lineage summary.

#### Scenario: Receipt contains policy basis and lineage
- **WHEN** an exportability decision is made
- **THEN** the resulting receipt SHALL include `policy_code`, `explanation`, and lineage rows with source class, asset identifier or label, and applied rule

#### Scenario: Draft and final stages are distinct
- **WHEN** exportability is evaluated for a draft artifact and again before final export
- **THEN** the receipt SHALL record whether the decision was made at `draft` or `final` stage
