## Why

The blockchain payment system (internal/payment/, internal/wallet/, internal/x402/) is fully implemented at the backend level, but users have no CLI commands to check balances, view transaction history, inspect spending limits, or send payments without going through the agent. This gap means payment features are invisible to direct user interaction.

## What Changes

- Add `lango payment` CLI command group with five subcommands: `balance`, `history`, `limits`, `info`, `send`
- Follow the existing `bootLoader` pattern (same as `lango security`) for lazy bootstrap initialization
- Register payment commands in `cmd/lango/main.go`
- Update README.md with payment CLI documentation, configuration reference, architecture tree, and a new "Blockchain Payments" section

## Capabilities

### New Capabilities
- `cli-payment-management`: CLI commands for viewing and managing blockchain payment operations (balance, history, limits, info, send)

### Modified Capabilities
- `payment-tools`: Executor agent tool list updated to include `payment_*` tools in multi-agent orchestration documentation

## Impact

- **New package**: `internal/cli/payment/` (6 files: payment.go, balance.go, history.go, limits.go, info.go, send.go)
- **Modified file**: `cmd/lango/main.go` — new import and command registration
- **Modified file**: `README.md` — features list, CLI commands, architecture tree, configuration reference, new Blockchain Payments section, wallet key security subsection
- **Dependencies**: Reuses existing `internal/payment.Service`, `internal/wallet.SpendingLimiter`, `internal/bootstrap.Result`
