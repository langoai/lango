## ADDED Requirements

### Requirement: Dynamic price quotes
The system SHALL compute price quotes for tools using base prices, rule evaluation, and optional trust/volume discounts. Quotes SHALL include basePrice, finalPrice, currency, modifiers, and validity period.

#### Scenario: Quote with base price
- **WHEN** Quote is called for a tool with base price 1000000
- **THEN** a Quote is returned with basePrice=1000000 and finalPrice reflecting any applicable discounts

#### Scenario: Quote for unpriced tool
- **WHEN** Quote is called for a tool with no base price set
- **THEN** the Quote is marked as isFree=true

### Requirement: Trust-based discounts
The system SHALL apply trust discounts when the peer's trust score exceeds 0.8. The discount percentage SHALL be configurable (default 10%).

#### Scenario: High trust peer discount
- **WHEN** Quote is called for peer with trust=0.9 and trustDiscount=0.10
- **THEN** finalPrice is reduced by 10% from basePrice

### Requirement: Paygate adapter
The system SHALL provide AdaptToPricingFunc() that returns a function compatible with paygate.PricingFunc signature: `func(toolName string) (price string, isFree bool)`.

#### Scenario: Adapter returns price string
- **WHEN** AdaptToPricingFunc() is called and the returned function is invoked with a priced tool
- **THEN** the price is returned as a USDC decimal string (e.g. "1.50")

### Requirement: Rule-based evaluation
The system SHALL support a RuleSet with ordered PricingRules that apply conditions and modifiers to base prices.

#### Scenario: Rule with condition match
- **WHEN** a rule condition matches the tool name
- **THEN** the rule's modifier is applied to the price
