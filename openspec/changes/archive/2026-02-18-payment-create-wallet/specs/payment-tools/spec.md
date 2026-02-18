## ADDED Requirements

### Requirement: Payment create wallet tool
The system SHALL provide a `payment_create_wallet` tool with SafetyLevel Dangerous that generates a new blockchain wallet. It MUST accept no required parameters and return the created wallet's address, chainId, and network name. If a wallet already exists, it MUST return the existing address with an informational message instead of creating a new one.

#### Scenario: Agent creates new wallet
- **WHEN** the agent calls `payment_create_wallet` and no wallet exists
- **THEN** a new wallet is created and the response contains `address`, `chainId`, `network`, and `status: "created"`

#### Scenario: Agent attempts to create wallet when one exists
- **WHEN** the agent calls `payment_create_wallet` and a wallet already exists
- **THEN** the response contains the existing `address`, `chainId`, `network`, and `status: "exists"` with a message indicating the wallet already exists

 #### Scenario: Wallet creation requires approval
- **WHEN** the agent calls `payment_create_wallet`
- **THEN** the tool's SafetyLevel is Dangerous, requiring user approval before execution
