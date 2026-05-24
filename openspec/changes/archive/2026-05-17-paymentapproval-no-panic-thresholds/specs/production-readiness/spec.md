## ADDED Requirements

### Requirement: Payment approval avoids production panic paths

The `internal/paymentapproval` package SHALL NOT contain production `panic` calls for deterministic threshold initialization or evaluator logic.

#### Scenario: Payment approval threshold setup does not panic
- **WHEN** the upfront-payment evaluator initializes its amount classification thresholds
- **THEN** threshold setup SHALL use deterministic non-panicking construction
- **AND** invalid runtime payment inputs SHALL continue to return reject outcomes instead of panicking
