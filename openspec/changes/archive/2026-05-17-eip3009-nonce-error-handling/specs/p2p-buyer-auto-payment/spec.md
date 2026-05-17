## MODIFIED Requirements

### Requirement: EIP-3009 authorization signing

The tool SHALL create and sign an EIP-3009 `transferWithAuthorization` using the buyer's wallet, the seller's address from the price quote, and the canonical USDC contract for the chain. If unsigned authorization creation fails, including secure nonce generation failure, the tool SHALL return an error and SHALL NOT sign or invoke the paid remote tool.

#### Scenario: Successful authorization signing
- **WHEN** the auto-approval check passes
- **THEN** the tool SHALL call `eip3009.NewUnsigned()` with the buyer address, seller address, amount, and a 10-minute deadline, then sign it with `eip3009.Sign()`

#### Scenario: Nonce generation failure aborts payment
- **WHEN** unsigned EIP-3009 authorization creation fails while generating the nonce
- **THEN** the paid invocation tool SHALL return an error identifying authorization creation
- **AND** it SHALL NOT call `eip3009.Sign()`
- **AND** it SHALL NOT invoke the paid remote tool
