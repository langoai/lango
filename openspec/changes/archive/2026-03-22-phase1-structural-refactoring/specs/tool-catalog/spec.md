## MODIFIED Requirements

### Requirement: Economy tools registered via domain package
Economy tools SHALL be registered via `economy.BuildTools()` called from the network module's `Init()` method, not via `app.buildEconomyTools()`. The `app/tools_economy.go` file SHALL be removed.

#### Scenario: Economy tools appear in catalog
- **WHEN** economy is enabled and engines are initialized
- **THEN** tools with names `economy_budget_allocate`, `economy_budget_status`, `economy_budget_close`, `economy_risk_assess`, `economy_negotiate`, `economy_negotiate_status`, `economy_escrow_create`, `economy_escrow_milestone`, `economy_escrow_status`, `economy_escrow_release`, `economy_escrow_dispute`, `economy_price_quote` are registered in the "economy" catalog category

#### Scenario: Excluded tools remain in app
- **WHEN** economy is enabled
- **THEN** `buildOnChainEscrowTools`, `buildSentinelTools`, `buildTeamEscrowTools` remain in `internal/app/` (not extracted in this change)
