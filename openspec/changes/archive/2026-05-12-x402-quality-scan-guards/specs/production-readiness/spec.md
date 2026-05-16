## ADDED Requirements

### Requirement: x402 static quality gates enforce dead-code and TODO hygiene

The `x402` package SHALL have executable tests that prevent `context.TODO()` reintroduction and legacy `NewX402Client` references.

#### Scenario: x402 package source contains no context.TODO calls
- **WHEN** the `x402` package quality guard tests scan `internal/x402` source files
- **THEN** the scan SHALL find zero `context.TODO()` occurrences

#### Scenario: Repository contains no legacy x402 client factory references
- **WHEN** the `x402` package quality guard tests scan repository Go files
- **THEN** the scan SHALL find zero `NewX402Client` references
