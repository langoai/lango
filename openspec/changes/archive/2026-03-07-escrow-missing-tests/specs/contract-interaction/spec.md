## MODIFIED Requirements

### Requirement: Contract caller provides read and write access
The contract package SHALL expose a `ContractCaller` interface with `Read` and `Write` methods that the concrete `Caller` struct implements. Consumers SHALL accept the interface type instead of the concrete struct.

#### Scenario: ContractCaller interface defined
- **WHEN** a package needs to call smart contracts
- **THEN** it SHALL depend on the `ContractCaller` interface, not the concrete `*Caller` struct

#### Scenario: Caller satisfies ContractCaller
- **WHEN** `*Caller` is used where `ContractCaller` is expected
- **THEN** it SHALL compile without error (compile-time interface check via `var _ ContractCaller = (*Caller)(nil)`)
