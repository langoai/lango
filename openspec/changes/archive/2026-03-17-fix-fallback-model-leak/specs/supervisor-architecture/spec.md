## MODIFIED Requirements

### Requirement: Fallback provider routing
The `ProviderProxy` SHALL reset `params.Model` to empty before calling the fallback provider, ensuring `Supervisor.Generate()` applies the fallback model name instead of carrying over the primary model.

#### Scenario: Primary fails and fallback receives correct model
- **WHEN** the primary provider fails and a fallback is configured
- **THEN** the fallback call SHALL have `params.Model == ""` so `Supervisor.Generate()` applies the fallback model

#### Scenario: Original params are not mutated
- **WHEN** a fallback call is made
- **THEN** the original `params` struct passed by the caller SHALL remain unchanged

#### Scenario: Primary succeeds without fallback
- **WHEN** the primary provider succeeds
- **THEN** the fallback provider SHALL NOT be called
