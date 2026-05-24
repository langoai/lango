## ADDED Requirements

### Requirement: Upfront payment approval tool
The meta tools surface SHALL provide an `approve_upfront_payment` tool that evaluates a prepayment request and records the result.

#### Scenario: Tool returns approval outcome
- **WHEN** `approve_upfront_payment` is invoked with transaction receipt ID, amount, trust input, and budget/policy context
- **THEN** it SHALL evaluate the request through the upfront payment approval domain model
- **AND** it SHALL return the decision, reason, suggested payment mode, amount class, and risk class

#### Scenario: Tool updates transaction receipt
- **WHEN** `approve_upfront_payment` completes
- **THEN** it SHALL update the linked transaction receipt with canonical payment approval state and append the corresponding event
