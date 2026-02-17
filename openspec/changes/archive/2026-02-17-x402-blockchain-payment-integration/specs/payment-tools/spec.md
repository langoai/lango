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
