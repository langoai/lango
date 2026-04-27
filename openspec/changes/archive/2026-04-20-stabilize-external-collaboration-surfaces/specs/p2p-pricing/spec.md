## MODIFIED Requirements

### Requirement: TrustThresholds config field
`P2PPricingConfig` SHALL include a `TrustThresholds` field with `PostPayMinScore` (float64, default 0.8), and the payment-side runtime SHALL treat that threshold inclusively for post-pay eligibility.

#### Scenario: Default trust threshold
- **WHEN** `TrustThresholds.PostPayMinScore` is zero or unset
- **THEN** the payment gate SHALL use 0.8 as the default threshold

#### Scenario: Exact threshold is post-pay eligible
- **WHEN** a peer reputation score is exactly equal to `postPayMinScore`
- **THEN** the payment-side runtime SHALL treat the request as post-pay eligible

## ADDED Requirements

### Requirement: Provider-side quote surface remains distinct from local policy pricing
The `p2p.pricing` surface SHALL remain the provider-side public quote surface exposed to remote peers. It SHALL NOT, by itself, imply that dynamic pricing, negotiation, or escrow policy engines are enabled.

#### Scenario: Provider-side quote semantics
- **WHEN** an operator configures `p2p.pricing`
- **THEN** the public P2P quote surface SHALL reflect provider-side quote behavior only
