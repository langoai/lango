## ADDED Requirements

### Requirement: EIP-2612 permit signing for USDC
The system SHALL provide an EIP-2612 permit signing builder that computes domain separators, struct hashes, and typed data hashes for USDC v2 permit operations.

#### Scenario: Compute domain separator
- **WHEN** given a chain ID and USDC contract address
- **THEN** the system SHALL return a 32-byte EIP-712 domain separator using USDC v2 domain parameters (name="USD Coin", version="2")

#### Scenario: Sign permit
- **WHEN** given owner, spender, value, nonce, deadline, chain ID, and USDC address
- **THEN** the system SHALL produce a valid EIP-2612 signature with V=27 or V=28, and R/S components that recover to the owner address

#### Scenario: Query permit nonce
- **WHEN** querying the USDC contract's `nonces(address)` function for an owner
- **THEN** the system SHALL return the current permit nonce as a big.Int

### Requirement: Circle permit paymaster provider
The system SHALL provide a `CirclePermitProvider` that builds `PaymasterAndData` using EIP-2612 permit signatures for Circle's on-chain paymaster contract.

#### Scenario: Stub mode for gas estimation
- **WHEN** `SponsorUserOp` is called with `stub=true`
- **THEN** the system SHALL return a `PaymasterAndData` of exactly 170 bytes (52-byte prefix + 118-byte zero-filled paymasterData) with the correct paymaster address in the first 20 bytes

#### Scenario: Real mode with permit signature
- **WHEN** `SponsorUserOp` is called with `stub=false`
- **THEN** the system SHALL query the permit nonce, sign an EIP-2612 permit, and return `PaymasterAndData` of exactly 170 bytes containing: paymaster(20) + verificationGas(16) + postOpGas(16) + mode(1) + token(20) + amount(32) + signature(65)

#### Scenario: Permit mode byte
- **WHEN** building paymasterData for permit mode
- **THEN** the mode byte SHALL be `0x01`

### Requirement: Permit paymaster config mode
The system SHALL support a `mode` field in `SmartAccountPaymasterConfig` with values `"rpc"` (default) and `"permit"`.

#### Scenario: Permit mode wiring
- **WHEN** config has `mode="permit"` and `provider="circle"`
- **THEN** the system SHALL create a `CirclePermitProvider` using the wallet provider and RPC client, without requiring a paymaster RPC URL

#### Scenario: RPC mode backward compatibility
- **WHEN** config has no `mode` field or `mode="rpc"`
- **THEN** the system SHALL create the existing RPC-based provider, requiring `rpcURL`
