## ADDED Requirements

### Requirement: x402 legacy client factory definition stays absent

The `x402` package SHALL have an executable test that prevents the legacy `NewX402Client` function definition from reappearing in package source.

#### Scenario: x402 package source contains no legacy client factory definition
- **WHEN** the x402 quality guard tests scan `internal/x402` source files
- **THEN** the scan SHALL find zero `NewX402Client` function definitions
