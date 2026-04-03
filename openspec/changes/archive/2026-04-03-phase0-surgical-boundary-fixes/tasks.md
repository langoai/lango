## 1. Shared Finance Package

- [x] 1.1 Add `github.com/shopspring/decimal` dependency to go.mod
- [x] 1.2 Create `internal/finance/usdc.go` with `ParseUSDC`, `FormatUSDC`, `FloatToMicroUSDC` using shopspring/decimal
- [x] 1.3 Create `internal/finance/quote.go` with `DefaultQuoteExpiry` constant
- [x] 1.4 Create `internal/finance/usdc_test.go` with table-driven tests for all functions
- [x] 1.5 Add wallet re-exports: `wallet.ParseUSDC`, `wallet.FormatUSDC`, `wallet.CurrencyUSDC`, `wallet.USDCDecimals` delegating to finance

## 2. Cross-Domain Boundary Fixes

- [x] 2.1 Create `internal/types/identity.go` with `DIDPrefix` constant
- [x] 2.2 Update `p2p/identity/identity.go` to use `types.DIDPrefix`
- [x] 2.3 Update `economy/escrow/address_resolver.go` to import `types` instead of `p2p/identity`
- [x] 2.4 Update `p2p/paygate/gate.go` to import `finance` instead of `wallet`
- [x] 2.5 Update `p2p/paygate/gate_test.go` to use `finance.CurrencyUSDC`

## 3. Type and Constant Unification

- [x] 3.1 Create `internal/types/reputation.go` with `ReputationQuerier` type
- [x] 3.2 Remove local `ReputationQuerier` from `economy/pricing/engine.go`, use `types.ReputationQuerier`
- [x] 3.3 Remove local `ReputationQuerier` from `economy/risk/engine.go`, use `types.ReputationQuerier`
- [x] 3.4 Replace `ReputationFunc` in `p2p/paygate/trust.go` with type alias to `types.ReputationQuerier`
- [x] 3.5 Simplify `app/wiring_economy.go` reputation wiring (remove type-casting adapter)
- [x] 3.6 Remove local `DefaultQuoteExpiry` from `p2p/paygate/gate.go`, use `finance.DefaultQuoteExpiry`
- [x] 3.7 Remove local `DefaultQuoteExpiry` from `economy/pricing/engine.go`, use `finance.DefaultQuoteExpiry`

## 4. Float Precision Bug Fixes

- [x] 4.1 Replace `app/convert.go` floatToMicroUSDC body with `finance.FloatToMicroUSDC`
- [x] 4.2 Replace `p2p/team/tools_escrow.go` inline `big.NewInt(int64(... * 1_000_000))` with `finance.FloatToMicroUSDC`
- [x] 4.3 Replace `economy/pricing/adapter.go` local `formatUSDC` and `USDCDecimals` with `finance.FormatUSDC` and `finance.USDCDecimals`

## 5. Frontmatter Render Unification

- [x] 5.1 Create `internal/mdparse/render.go` with `RenderFrontmatter` function
- [x] 5.2 Create `internal/mdparse/render_test.go` with roundtrip tests
- [x] 5.3 Update `agentregistry/parser.go` to use `mdparse.RenderFrontmatter`
- [x] 5.4 Update `skill/parser.go` to use `mdparse.RenderFrontmatter`

## 6. Architecture Boundary Enforcement

- [x] 6.1 Create `internal/archtest/boundary_test.go` with `go list`-based import graph analysis
- [x] 6.2 Implement report-only mode (`t.Log`) and verify 0 violations
- [x] 6.3 Switch to enforced mode (`t.Errorf`)
- [x] 6.4 Add depguard rules to `.golangci.yml` (economy↔p2p, p2p-infra↔wallet)
- [x] 6.5 Exclude p2p/handshake from depguard wallet restriction (with tracking comment)

## 7. Documentation

- [x] 7.1 Create `internal/memory/doc.go` with package role documentation
- [x] 7.2 Create `internal/agentmemory/doc.go` with package role documentation
- [x] 7.3 Create `internal/knowledge/doc.go` with package role documentation
- [x] 7.4 Create `internal/learning/doc.go` with package role documentation
- [x] 7.5 Create `docs/architecture/adr-001-package-boundaries.md`
- [x] 7.6 Create `docs/architecture/dependency-graph.md` with Mermaid diagram
