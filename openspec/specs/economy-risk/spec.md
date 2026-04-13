## Purpose

Capability spec for economy-risk. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Risk Assessor interface
The system SHALL provide an `Assessor` interface in `internal/economy/risk/` with method `Assess(ctx, peerDID, amount, verifiability)` that evaluates transaction risk using a 3-variable matrix (trust x value x verifiability) and returns an `Assessment` with recommended payment `Strategy`.

#### Scenario: Assess high-trust peer with low amount
- **WHEN** `Assess(ctx, peerDID, amount, verifiability)` is called with trust > 0.8
- **THEN** an Assessment is returned with Strategy="direct_pay" and RiskLevel="low"

#### Scenario: Assess unknown peer with high amount
- **WHEN** `Assess(ctx, peerDID, amount, verifiability)` is called with trust < 0.5 and amount > escrowThreshold
- **THEN** an Assessment is returned with Strategy="zk_escrow" and RiskLevel="critical"

### Requirement: 3-variable risk matrix (trust x value x verifiability)
The Assessor SHALL compute a `RiskScore` (0.0=safe to 1.0=risky) from three weighted factors:
- **Trust score**: from `reputation.Store` (weight ~0.4)
- **Transaction value**: relative to escrowThreshold (weight ~0.35)
- **Verifiability**: HIGH=0.0, MEDIUM=0.5, LOW=1.0 (weight ~0.25)

#### Scenario: Risk score calculation
- **WHEN** trust=0.9, amount=small, verifiability=HIGH
- **THEN** RiskScore is near 0.0 and RiskLevel="low"

#### Scenario: All factors adverse
- **WHEN** trust=0.1, amount=large, verifiability=LOW
- **THEN** RiskScore is near 1.0 and RiskLevel="critical"

### Requirement: Strategy selection rules
The Assessor SHALL select payment strategy based on the following decision matrix:

| Trust | Amount | Verifiability | Strategy |
|-------|--------|---------------|----------|
| > 0.8 | any | any | DirectPay |
| 0.5-0.8 | low | any | DirectPay or MicroPayment |
| 0.5-0.8 | high | any | Escrow |
| < 0.5 | low | HIGH | MicroPayment |
| < 0.5 | low | MEDIUM/LOW | ZKFirst |
| < 0.5 | high | any | ZKFirst + Escrow (zk_escrow) |
| any | > escrowThreshold | any | Escrow (forced) |

#### Scenario: High trust bypasses complexity
- **WHEN** trust > HighTrustScore (default 0.8)
- **THEN** Strategy is "direct_pay" regardless of amount or verifiability

#### Scenario: Medium trust with low amount
- **WHEN** trust is between MediumTrustScore (0.5) and HighTrustScore (0.8) and amount < escrowThreshold
- **THEN** Strategy is "direct_pay" or "micro_payment"

#### Scenario: Medium trust with high amount
- **WHEN** trust is between 0.5-0.8 and amount >= escrowThreshold
- **THEN** Strategy is "escrow"

#### Scenario: Low trust with low verifiable amount
- **WHEN** trust < 0.5 and amount < escrowThreshold and verifiability is HIGH
- **THEN** Strategy is "micro_payment"

#### Scenario: Low trust with unverifiable work
- **WHEN** trust < 0.5 and verifiability is LOW or MEDIUM
- **THEN** Strategy is "zk_first" (ZK proof required before payment)

#### Scenario: Low trust with high amount
- **WHEN** trust < 0.5 and amount >= escrowThreshold
- **THEN** Strategy is "zk_escrow" (ZK + escrow combined)

#### Scenario: Escrow forced for large amounts
- **WHEN** amount > RiskConfig.EscrowThreshold regardless of trust
- **THEN** Strategy includes escrow (either "escrow" or "zk_escrow")

### Requirement: RiskLevel classification
The system SHALL classify RiskScore into four levels:

| RiskScore Range | RiskLevel |
|-----------------|-----------|
| 0.0 - 0.25 | low |
| 0.25 - 0.50 | medium |
| 0.50 - 0.75 | high |
| 0.75 - 1.0 | critical |

#### Scenario: Borderline risk score
- **WHEN** RiskScore is exactly 0.50
- **THEN** RiskLevel is "high"

### Requirement: Assessment output
Each `Assess` call SHALL return an `Assessment` struct containing: PeerDID, Amount, TrustScore, Verifiability, RiskLevel, RiskScore, Strategy, Factors (list of weighted factors used), Explanation (human-readable), and AssessedAt timestamp.

#### Scenario: Assessment includes explanation
- **WHEN** an Assessment is generated
- **THEN** Explanation contains a human-readable description of why the strategy was chosen

#### Scenario: Assessment includes factors
- **WHEN** an Assessment is generated
- **THEN** Factors contains at least 3 entries: trust, value, verifiability with their values and weights

### Requirement: RiskConfig defaults
The system SHALL use the following defaults from `config.RiskConfig`:
- `EscrowThreshold`: "5.00" (USDC)
- `HighTrustScore`: 0.8
- `MediumTrustScore`: 0.5

#### Scenario: Default high trust threshold
- **WHEN** RiskConfig is not customized
- **THEN** peers with trust > 0.8 qualify for DirectPay

### Requirement: Integration with reputation.Store
The Assessor SHALL query `reputation.Store` (or equivalent trust provider) to obtain the current trust score for the given peerDID. If the peer has no reputation history, a default low-trust score (e.g., 0.3) SHALL be used.

#### Scenario: Unknown peer defaults to low trust
- **WHEN** the peerDID has no reputation records
- **THEN** TrustScore defaults to 0.3, resulting in conservative strategy selection

### Requirement: Verifiability enum
The system SHALL define three verifiability levels:
- `HIGH`: Output can be cryptographically verified (e.g., hash comparison, deterministic computation)
- `MEDIUM`: Output can be heuristically checked (e.g., LLM quality scoring)
- `LOW`: Output requires manual human review

#### Scenario: Verifiability affects strategy at low trust
- **WHEN** trust < 0.5 and verifiability is HIGH
- **THEN** MicroPayment is preferred over ZKFirst (lower overhead)
