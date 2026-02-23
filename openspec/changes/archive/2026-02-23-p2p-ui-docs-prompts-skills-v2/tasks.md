## 1. Core — Reputation GetDetails + Wiring

- [x] 1.1 Add `PeerDetails` struct and `GetDetails()` method to `internal/p2p/reputation/store.go`
- [x] 1.2 Add `pricingCfg config.P2PPricingConfig` field to `p2pComponents` struct in `internal/app/wiring.go`
- [x] 1.3 Wire `pricingCfg: cfg.P2P.Pricing` in `initP2P()` return statement

## 2. CLI — Reputation and Pricing Subcommands

- [x] 2.1 Create `internal/cli/p2p/reputation.go` with `newReputationCmd()` — table and JSON output
- [x] 2.2 Create `internal/cli/p2p/pricing.go` with `newPricingCmd()` — full list and tool filter modes
- [x] 2.3 Register `reputation` and `pricing` commands in `internal/cli/p2p/p2p.go`

## 3. Agent Tools — Price Query and Reputation

- [x] 3.1 Add `p2p_price_query` tool to `buildP2PTools()` in `internal/app/tools.go`
- [x] 3.2 Add `p2p_reputation` tool to `buildP2PTools()` in `internal/app/tools.go`

## 4. REST API — Reputation and Pricing Endpoints

- [x] 4.1 Add `p2pReputationHandler` to `internal/app/p2p_routes.go`
- [x] 4.2 Add `p2pPricingHandler` to `internal/app/p2p_routes.go`
- [x] 4.3 Register `/reputation` and `/pricing` routes in `registerP2PRoutes()`

## 5. Skills — New Definitions

- [x] 5.1 Create `skills/p2p-reputation/SKILL.md`
- [x] 5.2 Create `skills/p2p-pricing/SKILL.md`
- [x] 5.3 Create `skills/p2p-owner-shield/SKILL.md`

## 6. Prompts — Update Agent Guidance

- [x] 6.1 Update `prompts/AGENTS.md` P2P Network description with paid value exchange
- [x] 6.2 Update `prompts/TOOL_USAGE.md` with `p2p_price_query`, `p2p_reputation`, and paid workflow
- [x] 6.3 Update `prompts/agents/vault/IDENTITY.md` with reputation, pricing, and new REST endpoints

## 7. Documentation — Feature, Config, CLI Docs

- [x] 7.1 Add Paid Value Exchange section to `docs/features/p2p-network.md`
- [x] 7.2 Add Reputation System section to `docs/features/p2p-network.md`
- [x] 7.3 Add Owner Shield section to `docs/features/p2p-network.md`
- [x] 7.4 Add 9 config entries to `docs/configuration.md` (pricing, ownerProtection, minTrustScore)
- [x] 7.5 Add `reputation` and `pricing` command references to `docs/cli/p2p.md`

## 8. README — P2P Section Update

- [x] 8.1 Add Payment Gate, Reputation System, Owner Shield bullets to README.md P2P section
- [x] 8.2 Add Paid Value Exchange subsection with workflow steps

## 9. Example Configs

- [x] 9.1 Add pricing, ownerProtection, minTrustScore to `examples/p2p-trading/configs/alice.json`
- [x] 9.2 Add pricing, ownerProtection, minTrustScore to `examples/p2p-trading/configs/bob.json`
- [x] 9.3 Add pricing, ownerProtection, minTrustScore to `examples/p2p-trading/configs/charlie.json`

## 10. Verification

- [x] 10.1 Run `go build ./...` — confirm no compilation errors
- [x] 10.2 Run `go test ./internal/p2p/reputation/...` — confirm tests pass
