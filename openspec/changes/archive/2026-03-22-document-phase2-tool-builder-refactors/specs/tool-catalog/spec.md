## MODIFIED Requirements

### Requirement: Economy tools registered via domain package
Economy tools SHALL be registered via `economy.BuildTools()` called from the network module's `Init()` method. Sentinel tools SHALL be registered via `sentinel.BuildTools()` from the sentinel package. The `app/tools_economy.go` and `app/tools_sentinel.go` files SHALL NOT exist. Cycle-bound builders such as on-chain escrow and team-escrow MAY remain in `internal/app/`.

#### Scenario: Economy tools appear in catalog
- **WHEN** economy is enabled and engines are initialized
- **THEN** economy tools (budget, risk, negotiation, escrow, pricing) are registered in the "economy" catalog category via `economy.BuildTools()`

#### Scenario: Sentinel tools appear in catalog
- **WHEN** the sentinel engine is initialized
- **THEN** sentinel tools are registered in the "sentinel" catalog category via `sentinel.BuildTools()`

#### Scenario: Cycle-bound builders remain in app
- **WHEN** economy and team integrations are enabled
- **THEN** `buildOnChainEscrowTools` and `buildTeamEscrowTools` MAY remain in `internal/app/`
