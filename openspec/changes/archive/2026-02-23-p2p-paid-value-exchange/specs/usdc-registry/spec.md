## Purpose

Canonical USDC contract address registry for preventing fake token attacks in P2P payments.

## Requirements

### Requirement: Canonical Address Lookup
The system SHALL provide canonical USDC contract addresses for supported chains.

#### Scenario: Supported chain
- **WHEN** looking up the USDC address for Ethereum Mainnet (chain 1)
- **THEN** the system returns the canonical Circle USDC address

#### Scenario: Unsupported chain
- **WHEN** looking up the USDC address for an unsupported chain ID
- **THEN** the system returns an error

### Requirement: Address Verification
The system SHALL verify that a given address matches the canonical USDC address for a chain.

#### Scenario: Matching address
- **WHEN** checking if an address matches the canonical USDC for a chain
- **THEN** the system returns true for exact matches (case-insensitive)

#### Scenario: Non-matching address
- **WHEN** checking if a different address is canonical for a chain
- **THEN** the system returns false

### Requirement: On-Chain Verification
The system SHALL support on-chain verification of USDC contracts by checking symbol and decimals.
