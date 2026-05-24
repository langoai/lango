## ADDED Requirements

### Requirement: Artifact release approval tool
The meta tools surface SHALL provide an `approve_artifact_release` tool that evaluates a release request and returns a structured approval outcome.

#### Scenario: Tool returns structured approval outcome
- **WHEN** `approve_artifact_release` is invoked with artifact label, requested scope, and exportability state
- **THEN** it SHALL evaluate the request through the approval-flow domain model
- **AND** it SHALL return the decision, reason, issue classification, fulfillment assessment, and settlement hint

#### Scenario: Tool emits audit-backed approval receipt
- **WHEN** `approve_artifact_release` completes
- **THEN** it SHALL append an `artifact_release_approval` audit entry recording the approval outcome
