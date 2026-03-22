## ADDED Requirements

### Requirement: Domain packages own their tool builder functions
Domain packages SHALL export a `BuildTools()` function that creates agent tools from domain-owned exported types. The function MUST NOT import `internal/app/`.

#### Scenario: economy.BuildTools creates tools without app dependency
- **WHEN** `economy.BuildTools(be, re, ne, ee, pe)` is called with engine pointers
- **THEN** it returns the same tools previously built by `app.buildEconomyTools()`
- **AND** `internal/economy/tools.go` does not import `internal/app/`

#### Scenario: Nil engines are skipped gracefully
- **WHEN** `economy.BuildTools(nil, nil, nil, nil, nil)` is called
- **THEN** an empty tool slice is returned without panic

### Requirement: app/tools_economy.go is removed
The `buildEconomyTools` function and its 5 sub-builders SHALL be deleted from `internal/app/tools_economy.go`. Tool registration MUST go through `economy.BuildTools()` called from the network module's `Init()`.

#### Scenario: Module calls economy.BuildTools
- **WHEN** the network module initializes economy components
- **THEN** it calls `economy.BuildTools()` with individual engine pointers from `economyComponents`
