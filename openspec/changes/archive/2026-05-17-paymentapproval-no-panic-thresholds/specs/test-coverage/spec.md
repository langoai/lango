## ADDED Requirements

### Requirement: Payment approval panic guard stays executable

Executable tests SHALL prevent production `panic` calls from being reintroduced in `internal/paymentapproval` non-test Go files.

#### Scenario: Payment approval package panic regressions are rejected
- **WHEN** `internal/paymentapproval` non-test Go source files contain a `panic(` call
- **THEN** an executable package test SHALL fail and identify the offending file and line
