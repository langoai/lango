## ADDED Requirements

### Requirement: Risk assessment with 3-variable matrix
The system SHALL assess transaction risk using trust score, transaction amount, and output verifiability. The assessment SHALL produce a RiskLevel (low/medium/high/critical), RiskScore (0.0-1.0), and recommended Strategy.

#### Scenario: High trust peer with low amount
- **WHEN** Assess is called with trust=0.9, amount=100000, verifiability=high
- **THEN** RiskLevel is "low" and Strategy is "direct_pay"

#### Scenario: Low trust peer with high amount
- **WHEN** Assess is called with trust=0.3, amount=10000000, verifiability=low
- **THEN** RiskLevel is "high" or "critical" and Strategy includes escrow

#### Scenario: Amount exceeds escrow threshold
- **WHEN** Assess is called with amount exceeding configured escrow threshold
- **THEN** Strategy SHALL include "escrow" regardless of trust score

### Requirement: Strategy selection matrix
The system SHALL select payment strategies based on the following matrix:
- Trust > 0.8 → DirectPay
- Trust 0.5-0.8 + low amount → DirectPay or MicroPayment
- Trust 0.5-0.8 + high amount → Escrow
- Trust < 0.5 + low amount → MicroPayment or ZKFirst
- Trust < 0.5 + high amount → ZKFirst + Escrow
- Amount > escrowThreshold → Escrow (forced)

#### Scenario: Medium trust medium amount
- **WHEN** trust=0.6, amount=500000, verifiability=medium
- **THEN** Strategy is one of direct_pay, micro_payment, or escrow based on matrix

### Requirement: Reputation querier callback
The system SHALL use a ReputationQuerier function type to query trust scores, avoiding direct imports from the P2P reputation package.

#### Scenario: Reputation query failure
- **WHEN** ReputationQuerier returns an error
- **THEN** Assess returns the error without producing an assessment
