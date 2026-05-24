## ADDED Requirements

### Requirement: X402 fetch wrapper guards stay actionable

The `payment_x402_fetch` tool SHALL preserve actionable missing-parameter errors for its required wrapper inputs before network request construction begins.

#### Scenario: Missing X402 fetch URL fails at the wrapper
- **WHEN** `payment_x402_fetch` is invoked without `url`
- **THEN** the tool SHALL return an actionable missing-parameter error
