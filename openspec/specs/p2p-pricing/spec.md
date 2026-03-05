### Requirement: TrustThresholds config field
`P2PPricingConfig` SHALL include a `TrustThresholds` field with `PostPayMinScore` (float64, default 0.8).

#### Scenario: Default trust threshold
- **WHEN** `TrustThresholds.PostPayMinScore` is zero or unset
- **THEN** the payment gate uses 0.8 as the default threshold

### Requirement: SettlementConfig config field
`P2PPricingConfig` SHALL include a `Settlement` field with `ReceiptTimeout` (duration, default 2m) and `MaxRetries` (int, default 3).

#### Scenario: Default settlement config
- **WHEN** `Settlement.ReceiptTimeout` is zero
- **THEN** the settlement service uses 2 minutes as the default timeout
