## Context

The blockchain payment system is fully implemented at the core and application layers (`internal/payment/`, `internal/wallet/`, `internal/x402/`), with agent tools (`internal/tools/payment/`) providing access through the AI agent. However, there are no CLI commands for direct user interaction with the payment system — unlike security, memory, graph, and agent commands which all have CLI counterparts.

The existing CLI patterns (`internal/cli/security/`, `internal/cli/graph/`, etc.) provide a well-established template: a `bootLoader` function for lazy bootstrap, `--json` flag for machine-readable output, and tabwriter for human-readable tables.

## Goals / Non-Goals

**Goals:**
- Provide CLI commands for all read-only payment operations (balance, history, limits, info)
- Provide a CLI command for sending USDC with confirmation prompt
- Follow existing CLI patterns exactly (bootLoader, --json, tabwriter)
- Update README to document all payment features comprehensively

**Non-Goals:**
- No new business logic — CLI calls existing `payment.Service` methods directly
- No TUI onboard wizard integration for payment configuration (future work)
- No wallet key generation CLI — key management stays in the secrets store

## Decisions

### 1. bootLoader pattern with `*bootstrap.Result` (not `*config.Config`)

Payment commands need `*ent.Client`, `CryptoProvider`, and `*config.Config` — the full bootstrap result. Unlike memory/graph/agent commands that only need config, payment follows the security CLI pattern with `func() (*bootstrap.Result, error)`.

**Alternative**: Separate config + DB loaders. Rejected — adds complexity without benefit since bootstrap.Run provides everything in one call.

### 2. CLI-specific `initPaymentDeps` instead of reusing `wiring.go`

`wiring.go:initPayment()` returns nil on failure (graceful degradation for server mode). CLI commands need explicit errors so users know what's wrong. The CLI version returns `error` instead of nil, with actionable messages.

**Alternative**: Add an error-returning mode to wiring.go. Rejected — wiring.go serves server orchestration; CLI has different error semantics.

### 3. EntSpendingLimiter direct type (not SpendingLimiter interface)

`paymentDeps.limiter` is `*wallet.EntSpendingLimiter` (concrete) rather than `wallet.SpendingLimiter` (interface), because the `limits` command needs `MaxPerTx()` and `MaxDaily()` methods that are only on the concrete type. This mirrors the agent tool pattern in `internal/tools/payment/payment.go:163`.

### 4. Confirmation prompt for send command

The `send` command requires interactive confirmation by default, with `--force` for non-interactive use. This follows the existing pattern from `configDeleteCmd` and `newSecretsDeleteCmd`.

## Risks / Trade-offs

- **RPC connection per command**: Each CLI invocation creates a fresh RPC connection. This is acceptable for CLI usage but would be wasteful for batch operations. → Mitigation: CLI commands are infrequent; connection overhead is negligible for one-shot use.
- **No offline mode**: All payment commands require RPC connectivity. → Mitigation: Error messages clearly indicate RPC connection failure.
- **Duplicate initialization logic**: `initPaymentDeps` partially duplicates `wiring.go:initPayment()`. → Mitigation: Both are small (~40 lines); extracting a shared function would require changing wiring.go's nil-return semantics.
