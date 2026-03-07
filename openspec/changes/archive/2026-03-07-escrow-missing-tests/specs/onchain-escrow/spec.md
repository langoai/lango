## MODIFIED Requirements

### Requirement: Hub package clients accept ContractCaller interface
HubClient, VaultClient, FactoryClient, HubSettler, and VaultSettler constructors SHALL accept `contract.ContractCaller` interface instead of `*contract.Caller`.

#### Scenario: Constructors accept interface
- **WHEN** `NewHubClient`, `NewVaultClient`, `NewFactoryClient`, `NewHubSettler`, or `NewVaultSettler` is called
- **THEN** the `caller` parameter type SHALL be `contract.ContractCaller`

#### Scenario: Existing callers unaffected
- **WHEN** existing code passes `*contract.Caller` to hub package constructors
- **THEN** it SHALL compile without changes because `*Caller` satisfies `ContractCaller`
