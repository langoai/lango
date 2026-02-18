## ADDED Requirements

### Requirement: Payment send tool
The system SHALL provide a `payment_send` tool with SafetyLevel Dangerous that accepts `to` (address), `amount` (USDC string), and `purpose` (string) parameters. It MUST return txHash, status, amount, from, to, chainId, and network.

#### Scenario: Agent sends payment
- **WHEN** the agent calls `payment_send` with valid parameters
- **THEN** the payment is submitted and a receipt with txHash is returned

### Requirement: Payment balance tool
The system SHALL provide a `payment_balance` tool with SafetyLevel Safe that returns the wallet's USDC balance, address, chainId, and network name.

#### Scenario: Agent checks balance
- **WHEN** the agent calls `payment_balance`
- **THEN** the current USDC balance and wallet info are returned

### Requirement: Payment history tool
The system SHALL provide a `payment_history` tool with SafetyLevel Safe that accepts an optional `limit` parameter and returns recent transaction records.

#### Scenario: Agent views history
- **WHEN** the agent calls `payment_history` with limit 10
- **THEN** up to 10 most recent transactions are returned

### Requirement: Payment limits tool
The system SHALL provide a `payment_limits` tool with SafetyLevel Safe that returns maxPerTx, maxDaily, dailySpent, and dailyRemaining.

#### Scenario: Agent checks limits
- **WHEN** the agent calls `payment_limits`
- **THEN** the spending limits and current daily usage are returned

### Requirement: Payment wallet info tool
The system SHALL provide a `payment_wallet_info` tool with SafetyLevel Safe that returns the wallet address, chainId, and network name.

#### Scenario: Agent checks wallet info
- **WHEN** the agent calls `payment_wallet_info`
- **THEN** the wallet address and network information are returned

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
