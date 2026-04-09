## Purpose

Capability spec for p2p-team-payment. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Payment negotiation
The `p2p/team` package SHALL provide `NegotiatePayment` functionality that negotiates payment terms with remote agents before task delegation.

#### Scenario: Free tier for high-trust agents
- **WHEN** an agent has trust score > 0.9 and offers free tier
- **THEN** the payment mode SHALL be Free

#### Scenario: PostPay for trusted agents
- **WHEN** an agent has trust score > 0.7
- **THEN** the payment mode SHALL be PostPay (pay after task completion)

#### Scenario: PrePay for low-trust agents
- **WHEN** an agent has trust score <= 0.7
- **THEN** the payment mode SHALL be PrePay (pay before task execution)

### Requirement: PaymentAgreement type
The package SHALL define a `PaymentAgreement` struct with Mode (Free/PrePay/PostPay), Amount, Currency, TaskID, and AgentDID.

#### Scenario: Agreement tracks task
- **WHEN** a PaymentAgreement is created
- **THEN** it SHALL reference the specific TaskID and AgentDID

### Requirement: Budget validation
Payment negotiation SHALL validate that the requested amount does not exceed the configured budget limit before agreeing to payment.

#### Scenario: Over-budget rejection
- **WHEN** an agent requests payment exceeding the budget
- **THEN** the negotiation SHALL fail with a budget exceeded error

### Requirement: Integration with existing payment services
Payment execution SHALL use the existing `paygate.Gate` for payment authorization and `settlement.Service` for settlement. No new payment infrastructure SHALL be created.

#### Scenario: PayGate authorization
- **WHEN** a PrePay payment is authorized
- **THEN** it SHALL go through the existing PayGate authorization flow

### Requirement: Team budget bridge goroutine lifecycle
The `wireTeamBudgetBridge` function SHALL accept a `context.Context` parameter. Budget reservation timeout goroutines SHALL use a `select` on both the timer channel and `ctx.Done()` to ensure cleanup on application shutdown.

#### Scenario: Normal timeout releases reservation
- **WHEN** a budget reservation is made and the 5-minute timer expires before shutdown
- **THEN** the goroutine SHALL call `releaseFn()` via the timer path

#### Scenario: Shutdown cancels pending reservations
- **WHEN** the application context is cancelled (shutdown) before the 5-minute timer expires
- **THEN** the goroutine SHALL call `releaseFn()` via the `ctx.Done()` path, preventing access to a closed store
