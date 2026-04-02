## Why

The P2P and commerce domains had unnecessary coupling (economy importing p2p/identity, paygate importing wallet for key management), shared financial types were missing (causing duplicated constants and float64 precision bugs), and no automated boundary enforcement existed. Rather than a full bounded-context restructuring (68→7 packages), surgical evidence-based fixes deliver maximum value at minimum blast radius.

## What Changes

- New `internal/finance` package: extracts `ParseUSDC`, `FormatUSDC`, `FloatToMicroUSDC`, `CurrencyUSDC`, `USDCDecimals`, `DefaultQuoteExpiry` from wallet/economy. Uses shopspring/decimal for precise arithmetic
- Remove `economy/escrow` → `p2p/identity` import: move `DIDPrefix` constant to `types/identity.go`
- Remove `p2p/paygate` → `wallet` import: switch to `finance` package directly
- Unify `ReputationQuerier` (3 duplicate definitions) → `types/reputation.go`
- Unify `DefaultQuoteExpiry` (2 duplicate definitions) → `finance/quote.go`
- Fix float64 precision bug at 5 call sites → use `finance.FloatToMicroUSDC`
- Unify mdparse frontmatter render pattern → extract `mdparse.RenderFrontmatter`
- New `internal/archtest/boundary_test.go`: `go list`-based import graph boundary enforcement
- Add depguard rules to `.golangci.yml`: block economy↔p2p and p2p-infra↔wallet imports
- Add doc.go for knowledge subsystem packages (memory, agentmemory, knowledge, learning)
- Write ADR-001 package boundary policy document
- Generate package dependency Mermaid diagram

## Capabilities

### New Capabilities
- `shared-finance-types`: shared monetary types and utilities extracted from wallet into `internal/finance`
- `architecture-boundary-enforcement`: archtest + depguard-based automated package boundary verification

### Modified Capabilities
- `shared-types`: add `DIDPrefix` and `ReputationQuerier` to existing types package
- `shared-mdparse`: add `RenderFrontmatter` to existing mdparse package
- `domain-tool-builders`: explicitly excluded Unit 5 (team escrow tools move) to respect spec contract

## Impact

- **New packages**: `internal/finance`, `internal/archtest`
- **New files**: `types/identity.go`, `types/reputation.go`, `mdparse/render.go`, 4 doc.go files, ADR-001, dependency-graph.md
- **Modified files**: ~30 (import replacements, duplicate removal, re-export additions)
- **New dependency**: `github.com/shopspring/decimal v1.4.0`
- **Compatibility**: public API signatures unchanged, wallet re-exports maintain backward compatibility
