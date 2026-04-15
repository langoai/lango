## ADDED Requirements

### Requirement: Model pricing registry
The system SHALL provide `internal/provider/pricing.go` with a `ModelPrice{InputPerMillion, OutputPerMillion float64}` struct and a `PriceFor(model string) (ModelPrice, bool)` function. The registry SHALL include static entries for the primary supported models (Claude Opus, Claude Sonnet, Claude Haiku, Gemini 2.5 Pro, GPT-4o or equivalents). Unknown models SHALL return `(ModelPrice{}, false)`.

#### Scenario: Known model returns price
- **WHEN** `PriceFor("claude-opus-4-6")` is called
- **THEN** a non-zero `ModelPrice` and `true` SHALL be returned

#### Scenario: Unknown model returns false
- **WHEN** `PriceFor("nonexistent-model")` is called
- **THEN** `(ModelPrice{}, false)` SHALL be returned

### Requirement: Estimated cost on token usage message
`TurnTokenUsageMsg` SHALL gain a field `EstimatedCostUSD float64`. When the model's price is known, the field SHALL be computed as `(InputTokens × InputPerMillion + OutputTokens × OutputPerMillion) / 1_000_000`. When the price is unknown, the field SHALL be 0.

#### Scenario: Known model attaches cost
- **WHEN** a turn completes with 1000 input tokens, 500 output tokens on a priced model costing $15/M input and $75/M output
- **THEN** `EstimatedCostUSD` SHALL be `(1000 * 15 + 500 * 75) / 1_000_000 = 0.0525`

#### Scenario: Unknown model reports zero cost
- **WHEN** a turn completes on a model not in the pricing registry
- **THEN** `EstimatedCostUSD` SHALL be 0

### Requirement: Eventbus token event extended with cost
The existing eventbus token usage event SHALL gain an `EstimatedCostUSD` field populated identically to `TurnTokenUsageMsg`. A separate `CostEvent` SHALL NOT be introduced.

#### Scenario: Eventbus subscriber receives cost
- **WHEN** an eventbus subscriber receives a token usage event for a priced model turn
- **THEN** the event SHALL contain `EstimatedCostUSD` with the computed value

### Requirement: TUI renders cost next to token summary
The TUI token summary line SHALL display the estimated cost alongside token counts when `EstimatedCostUSD > 0`, formatted as `<input> in / <output> out / ~$<cost>` (two-decimal millicent precision or appropriate). When cost is 0 (unknown model), the cost suffix SHALL be omitted.

#### Scenario: Priced turn shows cost
- **WHEN** a turn with `EstimatedCostUSD=0.003` completes
- **THEN** the token summary line SHALL contain `~$0.003` (or similar formatting)

#### Scenario: Unpriced turn hides cost
- **WHEN** a turn with `EstimatedCostUSD=0` completes
- **THEN** the token summary line SHALL omit any cost suffix

### Requirement: /cost slash command
The TUI SHALL support a `/cost` slash command that prints the session's cumulative input tokens, output tokens, and estimated cost based on summing `TurnTokenUsageMsg` values observed during the session.

#### Scenario: /cost summarizes session
- **WHEN** the user types `/cost` after several turns
- **THEN** the TUI SHALL print a summary with total input tokens, total output tokens, and total estimated cost in USD
