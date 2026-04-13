## Context

Lango is a modular monolith with 68+ packages under `internal/`. A comprehensive restructuring proposal (7 bounded contexts) was reviewed and rejected in favor of surgical boundary fixes. The existing `app/wiring_*.go` (28 files) + `appinit` module system already provides domain separation. The 90-day roadmap rates build modularization as P4 (lowest priority).

## Goals / Non-Goals

**Goals:**
- Remove unnecessary import dependencies between P2P and commerce domains
- Extract shared financial types from wallet (key management) into an independent leaf package
- Fix float64→USDC precision bugs across all call sites
- Unify duplicated types/constants (ReputationQuerier ×3, DefaultQuoteExpiry ×2)
- Enforce package boundaries automatically in CI (archtest + depguard)

**Non-Goals:**
- Full package directory restructuring (deferred per ADR-001)
- Removing p2p/handshake → wallet dependency (requires Signer interface extraction)
- Bulk migration of 60+ wallet.* callers (re-export maintains backward compatibility)
- Moving team.BuildEscrowTools (domain-tool-builders spec contract respected)

## Decisions

### 1. New `internal/finance` package (extracted from wallet)
- **Choice**: Separate leaf package for monetary utilities
- **Alternative**: Keep in wallet → paygate continues depending on key management package
- **Rationale**: paygate only needs payment verification, not key management. finance has zero internal dependencies

### 2. shopspring/decimal adoption
- **Choice**: Replace `big.Rat`/`big.Float` with `shopspring/decimal`
- **Alternative**: Keep `big.Rat` → requires rounding hacks, unclear intent
- **Rationale**: De facto standard Go decimal library. `NewFromFloat` + `Round(0)` eliminates the 0.5 rounding hack

### 3. Type placement: `types` vs `finance`
- **Choice**: `DIDPrefix`, `ReputationQuerier` → types. `ParseUSDC`, `CurrencyUSDC` → finance
- **Alternative**: Everything in types → types becomes bloated
- **Rationale**: Separate monetary logic (decimal arithmetic) from simple type definitions

### 4. Wallet re-export for backward compatibility
- **Choice**: `wallet.ParseUSDC` delegates to `finance.ParseUSDC`
- **Alternative**: Bulk-replace 60+ callers → excessive blast radius for Phase 0
- **Rationale**: Minimizes scope. Gradual migration tracked via Deprecated comments

### 5. Dual enforcement: archtest + depguard
- **Choice**: Go test (`go list`-based) + golangci-lint (depguard) in parallel
- **Alternative**: Only one → test catches in CI, lint catches in IDE. Complementary
- **Rationale**: Defense in depth at different development stages

### 6. p2p/handshake excluded from depguard
- **Choice**: Temporarily exempt from wallet import restriction
- **Alternative**: Force enforcement → lint errors on legitimate WalletProvider usage
- **Rationale**: Requires Signer interface extraction as prerequisite. Tracked via comment

## Risks / Trade-offs

- **[wallet re-export longevity]** → Plan gradual migration in follow-up. Deprecated annotations track intent
- **[shopspring/decimal dependency]** → Stable library (v1.4.0), widely used. Minimal go.sum growth
- **[p2p/handshake exception]** → Resolve when Signer interface is extracted. Documented in depguard config and ADR-001
- **[archtest execution time]** → `go list` call takes ~2s. Negligible CI impact
