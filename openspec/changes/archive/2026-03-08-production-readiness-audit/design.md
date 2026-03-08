## Context

The codebase has grown rapidly under an MVP mindset, accumulating production-readiness debt across 14 packages. A three-pronged audit (stubs, test gaps, broken flows) identified: one crash-on-startup bug (enclave provider), one unimplemented feature (Telegram download), dead code (`NewX402Client`), `context.TODO()` in production, and zero test coverage on 7 security-critical packages (wallet, security, payment, smartaccount, economy/risk, p2p/team, p2p/protocol).

## Goals / Non-Goals

**Goals:**
- Eliminate all runtime stubs that crash or return "not implemented" errors
- Remove dead code and `context.TODO()` from production paths
- Achieve meaningful test coverage on all security-critical packages (crypto, payments, wallets)
- All changes independently verifiable — no cross-WU dependencies

**Non-Goals:**
- Integration tests with real blockchain networks (all tests are mocked)
- Refactoring or architectural changes to existing working code
- Adding new features beyond fixing identified audit findings
- E2E or TUI-level testing (unit tests only)

## Decisions

### D1: Remove `NewX402Client` entirely vs. fixing it
**Decision**: Remove entirely.
**Rationale**: The function is dead code — never called from any Go source. `Interceptor.HTTPClient` already provides the same functionality with proper context propagation, spending limits, and caching. Keeping it would create maintenance burden and confusion.

### D2: Enclave provider — remove case vs. improve error
**Decision**: Remove the `case "enclave"` branch and let it fall through to `default` with an actionable error listing valid providers.
**Rationale**: A dedicated case for an unimplemented provider is misleading. The `default` branch already handles unknown providers — it just needed a better error message listing valid options.

### D3: Telegram download — HTTP client injection
**Decision**: Use `Config.HTTPClient` field (injectable, defaults to `http.DefaultClient`) with 30s `context.WithTimeout`.
**Rationale**: Enables test mocking via `httptest.NewServer` without requiring interface abstractions. The timeout prevents hanging downloads.

### D4: Test strategy — mocks vs. real dependencies
**Decision**: All tests use mocks/stubs. No real network calls, no real databases (except in-memory SQLite via enttest).
**Rationale**: Tests must be fast, deterministic, and CI-friendly. Real blockchain/network dependencies would make tests flaky and slow.

### D5: 14 independent work units
**Decision**: Decompose into 14 independent WUs with no cross-dependencies.
**Rationale**: Enables maximum parallelization. Each WU touches a distinct set of files, so there are no merge conflicts. Each can be verified independently with `go build` + `go test` + `go vet`.

## Risks / Trade-offs

- **[Mock fidelity]** → Mocks may not capture all real-world failure modes. Mitigation: Focus mocks on interface boundaries; add integration tests in a future phase.
- **[Telegram API changes]** → The download implementation assumes stable `file.Link()` URL format. Mitigation: The telebot library abstracts this; any changes would be caught by the library update.
- **[Test maintenance]** → 14 new test files increase maintenance surface. Mitigation: Tests follow project conventions (table-driven, `tests`/`tt`/`give`/`want`) for consistency and readability.
