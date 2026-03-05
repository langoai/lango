## ADDED Requirements

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

## REMOVED Requirements

### Requirement: Gate.SubmitOnChain method
**Reason**: Replaced by the event-driven `settlement.Service` which handles the full on-chain submission lifecycle including signing, retry, and confirmation.
**Migration**: Callers of `Gate.SubmitOnChain()` should wire the settlement service to the event bus instead. The handler now publishes `ToolExecutionPaidEvent` which triggers settlement automatically.
