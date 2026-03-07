## 1. Economy Config

- [x] 1.1 Create internal/config/types_economy.go with EconomyConfig, BudgetConfig, RiskConfig, EscrowConfig, NegotiationConfig, DynamicPricingConfig structs
- [x] 1.2 Add Economy field to main Config struct in config/types.go
- [x] 1.3 Add unit tests for config defaults and mapstructure binding

## 2. Budget Subsystem

- [x] 2.1 Create internal/economy/budget/types.go with TaskBudget, SpendEntry, BudgetStatus, BudgetReport types
- [x] 2.2 Create internal/economy/budget/store.go with Store interface and in-memory implementation
- [x] 2.3 Create internal/economy/budget/engine.go with Allocate, Check, Record, Reserve, Close methods
- [x] 2.4 Create internal/economy/budget/options.go with functional options (WithHardLimit, WithAlertCallback, WithThresholds)
- [x] 2.5 Add table-driven tests for budget engine (allocate, spend check, record, reserve/release, threshold alerts, close)

## 3. Risk Assessment Subsystem

- [x] 3.1 Create internal/economy/risk/types.go with RiskLevel, Assessment, Verifiability, ReputationQuerier types
- [x] 3.2 Create internal/economy/risk/engine.go with Assess method and 3-variable risk matrix
- [x] 3.3 Create internal/economy/risk/strategy.go with strategy selection matrix (DirectPay, MicroPayment, Escrow, ZKFirst)
- [x] 3.4 Add table-driven tests for risk assessment (high trust low amount, low trust high amount, escrow threshold)

## 4. Dynamic Pricing Subsystem

- [x] 4.1 Create internal/economy/pricing/types.go with Quote, PricingRule, PriceModifier types
- [x] 4.2 Create internal/economy/pricing/rule.go with RuleSet and ordered rule evaluation
- [x] 4.3 Create internal/economy/pricing/engine.go with Quote method, trust/volume discounts
- [x] 4.4 Create internal/economy/pricing/adapters.go with AdaptToPricingFunc() returning paygate.PricingFunc compatible function
- [x] 4.5 Add table-driven tests for pricing (base price, trust discount, volume discount, rule evaluation, adapter)

## 5. Negotiation Subsystem

- [x] 5.1 Create internal/economy/negotiation/types.go with NegotiationSession, Terms, Phase enum
- [x] 5.2 Create internal/economy/negotiation/messages.go with JSON serialization for P2P transport
- [x] 5.3 Create internal/economy/negotiation/engine.go with Propose, Counter, Accept, Reject, CheckExpiry methods
- [x] 5.4 Create internal/economy/negotiation/strategy.go with auto-negotiation midpoint strategy
- [x] 5.5 Add table-driven tests for negotiation (lifecycle, turn-based validation, expiry, auto-respond)

## 6. Escrow Subsystem

- [x] 6.1 Create internal/economy/escrow/types.go with EscrowEntry, Milestone, EscrowStatus types
- [x] 6.2 Create internal/economy/escrow/store.go with Store interface and in-memory implementation
- [x] 6.3 Create internal/economy/escrow/engine.go with Create, Fund, CompleteMilestone, Release, Dispute, CheckExpiry
- [x] 6.4 Create internal/economy/escrow/lifecycle.go with state machine transition validation
- [x] 6.5 Add table-driven tests for escrow (lifecycle states, milestone completion, dispute, expiry, settlement)

## 7. Event Bus Integration

- [x] 7.1 Create internal/eventbus/economy_events.go with 8 economy event types (BudgetAlert, BudgetExhausted, NegotiationStarted/Completed/Failed, EscrowCreated/Milestone/Released)

## 8. App Wiring

- [x] 8.1 Create internal/app/wiring_economy.go with initEconomy(), economyComponents struct, and cross-system callback wiring
- [x] 8.2 Add economy component fields (interface{}) to App struct in app/types.go
- [x] 8.3 Wire initEconomy() call in app.New() at step 5o
- [x] 8.4 Add RequestNegotiatePropose/Respond types and NegotiatePayload to p2p/protocol/messages.go
- [x] 8.5 Add SetNegotiator setter and negotiate routing to p2p/protocol/handler.go

## 9. Agent Tools

- [x] 9.1 Create internal/app/tools_economy.go with buildEconomyTools() returning 12 economy tools
- [x] 9.2 Register economy tools in app.New() under "economy" catalog category

## 10. CLI Commands

- [x] 10.1 Create internal/cli/economy/economy.go with NewEconomyCmd command group
- [x] 10.2 Create internal/cli/economy/budget.go with budget status subcommand
- [x] 10.3 Create internal/cli/economy/risk.go with risk status subcommand
- [x] 10.4 Create internal/cli/economy/pricing.go with pricing status subcommand
- [x] 10.5 Create internal/cli/economy/negotiate.go with negotiate status subcommand
- [x] 10.6 Create internal/cli/economy/escrow.go with escrow status subcommand
- [x] 10.7 Register economy command in cmd/lango/main.go under GroupID "infra"

## 11. Verification

- [x] 11.1 Run go build ./... and verify clean build
- [x] 11.2 Run go test ./internal/economy/... and verify all tests pass
- [x] 11.3 Verify lango economy --help shows all subcommands
