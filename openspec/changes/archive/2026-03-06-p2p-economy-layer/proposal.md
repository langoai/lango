## Why

Lango has production-grade P2P infrastructure (libp2p, DID, Noise, reputation, USDC settlement, X402, paygate) but lacks an upper layer for autonomous economic behavior. Agents need to independently assess risk, manage budgets, negotiate prices, and protect funds through escrow to operate in a decentralized marketplace.

## What Changes

- Add per-task budget allocation, tracking, threshold alerts, and hard limit enforcement
- Add 3-variable risk assessment (trust × value × verifiability) with strategy routing (DirectPay, MicroPayment, Escrow, ZKFirst)
- Add rule-based dynamic pricing engine with trust/volume discounts and paygate adapter
- Add P2P negotiation protocol with propose/counter/accept/reject flow and auto-negotiation
- Add milestone-based escrow lifecycle (create → fund → activate → milestone → release/dispute/refund/expire)
- Wire all 5 subsystems through app init, event bus, and P2P protocol handler
- Add 12 agent tools for runtime economy operations
- Add `lango economy` CLI command group with budget/risk/pricing/negotiate/escrow subcommands

## Capabilities

### New Capabilities
- `economy-budget`: Per-task budget allocation, spend tracking, threshold alerts, burn rate, hard limit enforcement
- `economy-risk`: Trust-based risk assessment with 3-variable matrix and payment strategy routing
- `economy-pricing`: Dynamic pricing engine with rule evaluation, trust/volume discounts, paygate adapter
- `economy-negotiation`: P2P price negotiation protocol with propose/counter/accept/reject state machine
- `economy-escrow`: Milestone-based escrow lifecycle with fund locking, milestone completion, dispute/refund
- `economy-wiring`: App integration layer wiring budget, risk, pricing, negotiation, escrow into event bus, P2P protocol, and tool catalog
- `economy-cli`: CLI commands for inspecting economy layer status and configuration

### Modified Capabilities
- `application-core`: Added economy component fields to App struct and initEconomy() wiring in app.New()
- `config-types`: Added EconomyConfig to root Config struct with budget/risk/negotiate/escrow/pricing sub-configs

## Impact

- **New packages**: `internal/economy/{budget,risk,pricing,negotiation,escrow}/` (25+ files)
- **Modified packages**: `internal/app/` (wiring, tools, types), `internal/p2p/protocol/` (negotiate messages/handler), `internal/eventbus/` (8 economy events), `internal/config/` (economy config types), `cmd/lango/` (CLI registration), `internal/cli/economy/` (CLI commands)
- **Dependencies**: Uses existing `internal/wallet` (ParseUSDC), `internal/p2p/reputation` (trust scores), `internal/p2p/paygate` (PricingFunc adapter)
- **Config**: New `economy.*` config namespace with 5 sub-configs
