## ADDED Requirements

### Requirement: X402 fetch tool uses payment prefix
The `payment_x402_fetch` tool SHALL use the `payment_` prefix in its tool name to ensure correct routing to the vault sub-agent via `PartitionTools()` prefix matching.

#### Scenario: Tool routes to vault agent
- **WHEN** `PartitionTools()` processes the `payment_x402_fetch` tool
- **THEN** the tool is assigned to the vault agent's tool set (not Unmatched)

#### Scenario: Tool name matches payment prefix convention
- **WHEN** listing all payment tools
- **THEN** `payment_x402_fetch` follows the same `payment_` prefix convention as `payment_send`, `payment_balance`, `payment_history`, `payment_limits`, `payment_wallet_info`, and `payment_create_wallet`
