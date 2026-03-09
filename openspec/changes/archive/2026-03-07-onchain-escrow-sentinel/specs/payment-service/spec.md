## MODIFIED Requirements

### Requirement: Escrow configuration
The EscrowConfig SHALL include an `OnChain` sub-struct (`EscrowOnChainConfig`) with fields: Enabled (bool), Mode (string: "hub"|"vault"), HubAddress, VaultFactoryAddress, VaultImplementation, ArbitratorAddress, TokenAddress (all string), and PollInterval (time.Duration). All fields SHALL have `mapstructure` and `json` struct tags. The default for Enabled SHALL be false, preserving backward compatibility.

#### Scenario: On-chain config disabled by default
- **WHEN** no `economy.escrow.onChain` section is present in config
- **THEN** EscrowOnChainConfig.Enabled defaults to false and custodian mode is used

#### Scenario: Hub mode config
- **WHEN** config sets `economy.escrow.onChain.enabled=true` and `mode=hub` with `hubAddress`
- **THEN** the system initializes HubSettler with the configured hub and token addresses
