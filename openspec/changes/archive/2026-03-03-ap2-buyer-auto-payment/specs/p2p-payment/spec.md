## MODIFIED Requirements

### Requirement: P2P payment tool registration
The system SHALL register P2P payment tools (`p2p_pay` and `p2p_invoke_paid`) when the payment service and wallet are available. `buildP2PPaymentTool` SHALL return the `p2p_pay` tool, and `buildP2PPaidInvokeTool` SHALL return the `p2p_invoke_paid` tool. Both tool sets SHALL be appended to the P2P tool list during initialization.

#### Scenario: Both payment tools registered
- **WHEN** the application initializes with `p2p.enabled=true` and valid payment components (wallet, limiter, service)
- **THEN** the P2P tool list SHALL include both `p2p_pay` and `p2p_invoke_paid`

#### Scenario: p2p_pay available without p2p_invoke_paid
- **WHEN** `paymentComponents` has a service but nil limiter
- **THEN** `buildP2PPaymentTool` SHALL return `p2p_pay` but `buildP2PPaidInvokeTool` SHALL return nil

#### Scenario: Tool unavailable without payment service
- **WHEN** the application is initialized with `payment.enabled=false`
- **THEN** `buildP2PPaymentTool` SHALL return nil and `p2p_pay` SHALL NOT be registered with the agent
- **AND** `buildP2PPaidInvokeTool` SHALL return nil and `p2p_invoke_paid` SHALL NOT be registered
