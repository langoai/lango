## Purpose

Capability spec for economy-pricing. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Rule-based dynamic pricing engine
The system SHALL provide a dynamic pricing engine in `internal/economy/pricing/` that computes `Quote` prices by applying an ordered `RuleSet` of `PricingRule` entries to a base price.

#### Scenario: Evaluate rules in priority order
- **WHEN** `RuleSet.Evaluate(toolName, trustScore, peerDID, basePrice)` is called
- **THEN** rules are evaluated in ascending priority order and all matching rules' modifiers are applied cumulatively

#### Scenario: No matching rules
- **WHEN** no rules match the given context
- **THEN** the base price is returned unchanged with no modifiers

### Requirement: PriceModifier types
The system SHALL support four modifier types:
- `trust_discount`: discount based on peer trust score (e.g., factor=0.9 for 10% discount)
- `volume_discount`: discount based on transaction history volume (e.g., factor=0.95 for 5% discount)
- `surge`: price increase during high demand (e.g., factor=1.2 for 20% markup)
- `custom`: arbitrary modifier with custom description

#### Scenario: Trust discount applied
- **WHEN** a rule with ModifierType="trust_discount" and Factor=0.9 matches
- **THEN** the price is multiplied by 0.9 (10% discount)

#### Scenario: Surge pricing applied
- **WHEN** a rule with ModifierType="surge" and Factor=1.5 matches
- **THEN** the price is multiplied by 1.5 (50% markup)

#### Scenario: Multiple modifiers stack
- **WHEN** two rules match (trust_discount factor=0.9, volume_discount factor=0.95)
- **THEN** the final price is basePrice * 0.9 * 0.95

### Requirement: PricingRule structure
Each PricingRule SHALL contain: Name (unique identifier), Priority (int, lower=higher priority), Condition (RuleCondition), Modifier (PriceModifier), and Enabled (bool).

#### Scenario: Disabled rule is skipped
- **WHEN** a rule has Enabled=false
- **THEN** it is not evaluated even if its condition would match

### Requirement: RuleCondition matching
A `RuleCondition` SHALL support filtering on:
- `ToolPattern`: glob pattern for tool name (e.g., "search_*", "compute_*")
- `MinTrustScore` / `MaxTrustScore`: trust score range
- `PeerDID`: specific peer targeting

All non-empty fields must match for the condition to be satisfied (AND logic).

#### Scenario: Tool pattern matching
- **WHEN** ToolPattern="search_*" and toolName="search_web"
- **THEN** the condition matches

#### Scenario: Trust score range matching
- **WHEN** MinTrustScore=0.5, MaxTrustScore=0.8, and trustScore=0.6
- **THEN** the condition matches

#### Scenario: Peer-specific rule
- **WHEN** PeerDID="did:lango:abc123" and the invoking peer matches
- **THEN** the condition matches only for that specific peer

#### Scenario: Empty condition matches everything
- **WHEN** all RuleCondition fields are zero-valued
- **THEN** the condition matches all requests

### Requirement: RuleSet management
The `RuleSet` SHALL provide `Add`, `Remove`, and `Rules` methods. Rules are kept sorted by priority (ascending) after each Add.

#### Scenario: Add rule maintains sort order
- **WHEN** rules with priorities [3, 1, 2] are added
- **THEN** `Rules()` returns them in order [1, 2, 3]

#### Scenario: Remove rule by name
- **WHEN** `Remove("old_rule")` is called
- **THEN** the rule with Name="old_rule" is removed from the set

#### Scenario: Rules returns a copy
- **WHEN** `Rules()` is called
- **THEN** a copy of the internal slice is returned (mutations do not affect the RuleSet)

### Requirement: Integer arithmetic for price calculation
The system SHALL use basis-point integer arithmetic (10000 = 1.0x) to apply modifiers, avoiding floating-point precision issues with USDC amounts.

#### Scenario: Basis-point multiplication
- **WHEN** price=1000000 (1 USDC) and factor=0.9
- **THEN** result = 1000000 * 9000 / 10000 = 900000 (0.90 USDC)

### Requirement: Quote output
The pricing engine SHALL produce a `Quote` containing: ToolName, BasePrice, FinalPrice, Currency ("USDC"), Modifiers (applied list), IsFree (bool for zero-cost tools), ValidUntil (quote expiry), and PeerDID.

#### Scenario: Free tool quote
- **WHEN** a tool has BasePrice=0
- **THEN** Quote.IsFree=true and FinalPrice=0

#### Scenario: Quote includes validity window
- **WHEN** a quote is generated
- **THEN** ValidUntil is set to a reasonable future time (e.g., 5 minutes)

### Requirement: DynamicPricingConfig defaults
The system SHALL use the following defaults from `config.DynamicPricingConfig`:
- `Enabled`: false (opt-in)
- `TrustDiscount`: 0.1 (max 10% discount for high-trust peers)
- `VolumeDiscount`: 0.05 (max 5% discount for high-volume peers)
- `MinPrice`: "0.01" (USDC floor)

#### Scenario: Price floor enforcement
- **WHEN** modifiers reduce the price below MinPrice
- **THEN** FinalPrice is clamped to MinPrice

### Requirement: AdaptToPricingFunc adapter
The system SHALL provide an `AdaptToPricingFunc()` function that converts the pricing engine into a `paygate.PricingFunc` compatible callback, allowing the paygate layer to query dynamic prices without direct dependency on the pricing package.

#### Scenario: PricingFunc adapter called by paygate
- **WHEN** paygate invokes the PricingFunc with toolName and peerDID
- **THEN** the pricing engine evaluates rules and returns the computed price
