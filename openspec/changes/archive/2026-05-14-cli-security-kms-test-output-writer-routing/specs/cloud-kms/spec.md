## ADDED Requirements

### Requirement: KMS test output routing

`lango security kms test` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture roundtrip progress and success output without intercepting process-global stdout.

#### Scenario: KMS roundtrip progress writes to command output
- **WHEN** `lango security kms test` is run
- **THEN** the command writes the roundtrip progress and success lines to the Cobra command output stream
