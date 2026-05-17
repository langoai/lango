## MODIFIED Requirements

### Requirement: Structured upfront payment approval
The system SHALL provide a structured upfront payment approval model for `knowledge exchange v1`. It SHALL emit one of `approve`, `reject`, or `escalate`. Invalid runtime amount or user max-prepay inputs SHALL return structured `reject` outcomes instead of panicking.

#### Scenario: Low-risk prepay is approved
- **WHEN** an upfront payment request is within current policy and budget limits and trust conditions are acceptable
- **THEN** the approval flow SHALL return `approve`

#### Scenario: Budget or policy failure rejects
- **WHEN** an upfront payment request violates budget or prepay policy
- **THEN** the approval flow SHALL return `reject`

#### Scenario: High amount or trust edge case escalates
- **WHEN** an upfront payment request crosses configured amount or risk thresholds or enters a low-trust edge case
- **THEN** the approval flow SHALL return `escalate`

#### Scenario: Invalid runtime amounts reject without panic
- **WHEN** an upfront payment request contains an invalid payment amount or user max-prepay value
- **THEN** the approval flow SHALL return `reject`
- **AND** it SHALL include a policy code describing the invalid input
- **AND** it SHALL NOT panic
