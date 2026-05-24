## ADDED Requirements

### Requirement: X402 fetch wrapper input guards stay actionable

The `payment_x402_fetch` tool SHALL reject a missing required `url` input with an actionable missing-parameter error before attempting to create an HTTP client or request.

#### Scenario: Missing URL returns a wrapper validation error
- **WHEN** the agent calls `payment_x402_fetch` without `url`
- **THEN** the tool SHALL return an actionable missing-parameter error
