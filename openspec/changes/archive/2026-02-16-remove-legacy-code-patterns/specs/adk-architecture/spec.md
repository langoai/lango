## MODIFIED Requirements

### Requirement: Event history truncation
The `EventsAdapter` SHALL use token-budget-based truncation for all cases. When no explicit token budget is provided, it SHALL use a default budget of 32000 tokens. The adapter SHALL include messages from most recent to oldest until the budget is exhausted.

#### Scenario: Explicit token budget
- **WHEN** a token budget is provided via `EventsWithTokenBudget(budget)`
- **THEN** messages SHALL be included from most recent until the budget is exhausted

#### Scenario: No token budget (default)
- **WHEN** no token budget is set (budget = 0)
- **THEN** the default budget of 32000 tokens SHALL be used

## REMOVED Requirements

### Requirement: Legacy 100-message hardcap
**Reason**: The hardcoded 100-message cap when `tokenBudget=0` is replaced by a default token budget of 32000. Token-based truncation is more accurate than message count.
**Migration**: No action needed; the default token budget provides equivalent or better behavior.
