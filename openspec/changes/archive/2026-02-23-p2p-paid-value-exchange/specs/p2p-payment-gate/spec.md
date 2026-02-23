## Purpose

Payment gate that sits between the P2P firewall and tool executor, enforcing USDC payment requirements for paid tool invocations using EIP-3009 pre-signed authorizations.

## Requirements

### Requirement: Price Query
The system SHALL allow remote peers to query tool pricing before invocation.

#### Scenario: Free tool query
- **WHEN** a peer queries the price of a tool with no configured price
- **THEN** the system returns isFree=true

#### Scenario: Paid tool query
- **WHEN** a peer queries the price of a tool with configured pricing
- **THEN** the system returns a PriceQuote containing toolName, price, currency, USDC contract, chainId, sellerAddr, and quoteExpiry

### Requirement: Payment Verification
The system SHALL verify EIP-3009 payment authorizations before executing paid tools.

#### Scenario: Valid authorization
- **WHEN** a paid tool invocation includes a valid EIP-3009 authorization with correct recipient, sufficient amount, and unexpired deadline
- **THEN** the system returns StatusVerified and proceeds with tool execution

#### Scenario: Missing authorization
- **WHEN** a paid tool invocation does not include paymentAuth
- **THEN** the system returns StatusPaymentRequired with a PriceQuote

#### Scenario: Insufficient payment
- **WHEN** the authorization value is less than the tool price
- **THEN** the system returns StatusInvalid with reason "insufficient payment"

#### Scenario: Expired authorization
- **WHEN** the authorization's validBefore is in the past
- **THEN** the system returns StatusInvalid with reason "payment authorization expired"

### Requirement: Canonical USDC Verification
The system SHALL verify that the USDC contract address matches the canonical address for the chain.

#### Scenario: Non-canonical contract
- **WHEN** the configured USDC contract does not match the canonical address for the chain
- **THEN** the system returns StatusInvalid with reason indicating non-canonical USDC contract
