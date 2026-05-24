## MODIFIED Requirements

### Requirement: Payment negotiation
The `p2p/team` package SHALL provide `NegotiatePayment` functionality that negotiates payment terms with remote agents before task delegation.

#### Scenario: Free tier for high-trust agents
- **WHEN** an agent has trust score > 0.9 and offers free tier
- **THEN** the payment mode SHALL be Free

#### Scenario: PostPay for trusted agents
- **WHEN** an agent has trust score >= 0.8
- **THEN** the payment mode SHALL be PostPay (pay after task completion)

#### Scenario: PrePay for low-trust agents
- **WHEN** an agent has trust score < 0.8
- **THEN** the payment mode SHALL be PrePay (pay before task execution)
