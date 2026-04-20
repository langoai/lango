## Purpose

Capability spec for approval-flow. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Structured artifact release approval
The system SHALL provide a structured artifact release approval model for `knowledge exchange v1`. It SHALL emit one of `approve`, `reject`, `request-revision`, or `escalate`.

#### Scenario: Exportable scope-matched artifact is approved
- **WHEN** an artifact release request has matching requested artifact label and an exportability decision state of `exportable`
- **THEN** the approval flow SHALL return `approve`

#### Scenario: Scope mismatch requests revision
- **WHEN** an artifact release request is otherwise valid but the submitted artifact does not match the requested artifact label
- **THEN** the approval flow SHALL return `request-revision`

#### Scenario: Needs-human-review escalates
- **WHEN** the exportability decision state is `needs-human-review`
- **THEN** the approval flow SHALL return `escalate`

#### Scenario: Blocked artifact without override is rejected
- **WHEN** the exportability decision state is `blocked` and no override is requested
- **THEN** the approval flow SHALL return `reject`

### Requirement: Structured release outcome records
Every non-approve release decision SHALL include structured outcome data covering reason, issue classification, fulfillment assessment, and settlement hint.

#### Scenario: Revision contains scope issue
- **WHEN** a release decision is `request-revision`
- **THEN** the outcome SHALL include an issue classification such as `scope_mismatch`

#### Scenario: Reject contains settlement hint
- **WHEN** a release decision is `reject`
- **THEN** the outcome SHALL include a settlement hint suitable for later economic handling
