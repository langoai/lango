## MODIFIED Requirements

### Requirement: Token budget wired at construction
The wiring layer SHALL pass `ModelTokenBudget(cfg.Agent.Model)` to the agent constructor via `WithAgentTokenBudget` option. This sets the session history token budget on `SessionServiceAdapter` before the runner is created.

#### Scenario: Token budget derived from model
- **WHEN** agent is constructed with model name "claude-3.5-sonnet"
- **THEN** `SessionServiceAdapter.tokenBudget` SHALL be set to the value returned by `ModelTokenBudget("claude-3.5-sonnet")`
