## ADDED Requirements

### Requirement: Smart account module panic guard stays executable

Executable tests SHALL prevent production `panic` calls from being reintroduced in `internal/smartaccount/module` non-test Go files.

#### Scenario: Smart account module panic regressions are rejected
- **WHEN** `internal/smartaccount/module` non-test Go source files contain a `panic(` call
- **THEN** an executable package test SHALL fail and identify the offending file and line
