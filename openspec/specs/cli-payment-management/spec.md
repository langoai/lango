# CLI Payment Management

## Purpose

Provides CLI commands for managing blockchain payment operations, allowing users to view balances, transaction history, spending limits, wallet information, and send USDC payments directly from the terminal without going through the AI agent.

## Requirements

### Requirement: Payment command group
The system SHALL provide a `lango payment` command group that contains subcommands for managing blockchain payment operations.

#### Scenario: Payment help output
- **WHEN** user runs `lango payment --help`
- **THEN** the system SHALL display available subcommands: balance, history, limits, info, send

### Requirement: Balance command
The system SHALL provide a `lango payment balance` command that displays the wallet's USDC balance, address, and network.

#### Scenario: Display balance in text format
- **WHEN** user runs `lango payment balance`
- **THEN** the system SHALL display balance in USDC, wallet address, and network name with chain ID

#### Scenario: Display balance in JSON format
- **WHEN** user runs `lango payment balance --json`
- **THEN** the system SHALL output JSON with fields: balance, currency, address, chainId, network

#### Scenario: Payment disabled error
- **WHEN** user runs `lango payment balance` and `payment.enabled` is false
- **THEN** the system SHALL return an error indicating payment is not enabled

### Requirement: History command
The system SHALL provide a `lango payment history` command that displays recent payment transactions in a table.

#### Scenario: Display transaction history
- **WHEN** user runs `lango payment history`
- **THEN** the system SHALL display a table with columns: STATUS, AMOUNT, TO, PURPOSE, TX HASH, CREATED

#### Scenario: Limit history results
- **WHEN** user runs `lango payment history --limit 5`
- **THEN** the system SHALL display at most 5 transactions

#### Scenario: Empty history
- **WHEN** user runs `lango payment history` and no transactions exist
- **THEN** the system SHALL display "No transactions found."

#### Scenario: JSON history output
- **WHEN** user runs `lango payment history --json`
- **THEN** the system SHALL output JSON with fields: transactions (array), count

### Requirement: Limits command
The system SHALL provide a `lango payment limits` command that displays spending limits and daily usage.

#### Scenario: Display spending limits
- **WHEN** user runs `lango payment limits`
- **THEN** the system SHALL display max per transaction, max daily, spent today, and remaining today in USDC

#### Scenario: JSON limits output
- **WHEN** user runs `lango payment limits --json`
- **THEN** the system SHALL output JSON with fields: maxPerTx, maxDaily, dailySpent, dailyRemaining, currency

### Requirement: Info command
The system SHALL provide a `lango payment info` command that displays wallet and payment system configuration.

#### Scenario: Display payment info
- **WHEN** user runs `lango payment info`
- **THEN** the system SHALL display wallet address, network, wallet provider, USDC contract, RPC URL, and X402 status

#### Scenario: JSON info output
- **WHEN** user runs `lango payment info --json`
- **THEN** the system SHALL output JSON with fields: address, chainId, network, walletProvider, usdcContract, rpcUrl, x402

### Requirement: Send command
The system SHALL provide a `lango payment send` command that sends USDC to a recipient address with required flags --to, --amount, and --purpose.

#### Scenario: Interactive send with confirmation
- **WHEN** user runs `lango payment send --to 0x... --amount 1.00 --purpose "test"` in an interactive terminal
- **THEN** the system SHALL display a confirmation prompt showing amount, recipient, network, and purpose
- **AND** proceed only if the user confirms with "y" or "yes"

#### Scenario: Non-interactive send with force flag
- **WHEN** user runs `lango payment send --to 0x... --amount 1.00 --purpose "test" --force`
- **THEN** the system SHALL skip the confirmation prompt and send immediately

#### Scenario: Non-interactive without force flag
- **WHEN** user runs `lango payment send --to 0x... --amount 1.00 --purpose "test"` in a non-interactive terminal without --force
- **THEN** the system SHALL return an error indicating --force is required

#### Scenario: Missing required flags
- **WHEN** user runs `lango payment send` without --to, --amount, or --purpose
- **THEN** the system SHALL return an error indicating which flags are required

#### Scenario: Successful send output
- **WHEN** the payment is submitted successfully
- **THEN** the system SHALL display status, tx hash, amount, from, to, and network

### Requirement: Bootstrap error handling
The system SHALL return descriptive errors when payment dependencies cannot be initialized, rather than silently degrading.

#### Scenario: Payment not enabled
- **WHEN** `payment.enabled` is false
- **THEN** the system SHALL return an error: "payment system is not enabled (set payment.enabled = true)"

#### Scenario: RPC connection failure
- **WHEN** the RPC endpoint is unreachable
- **THEN** the system SHALL return an error indicating the RPC URL and the connection failure reason
