## MODIFIED Requirements

### Requirement: Escrow settlement executor selection
The escrow engine SHALL use `USDCSettler` as the `SettlementExecutor` when `paymentComponents` is available (payment system enabled). The escrow engine SHALL fall back to `noopSettler` when payment is not available. The `EscrowConfig` SHALL include a `Settlement` sub-config with `ReceiptTimeout` and `MaxRetries` fields.

#### Scenario: Payment enabled uses USDC settler
- **WHEN** the economy layer is initialized with non-nil `paymentComponents`
- **THEN** `USDCSettler` is created with the payment wallet, tx builder, and RPC client

#### Scenario: Payment disabled uses noop settler
- **WHEN** the economy layer is initialized with nil `paymentComponents`
- **THEN** `noopSettler` is used and escrow operations succeed without on-chain activity

#### Scenario: Settlement config applied to settler
- **WHEN** `EscrowConfig.Settlement.ReceiptTimeout` and `MaxRetries` are configured
- **THEN** the `USDCSettler` is created with those values via functional options
