## Why

Several destructive or setup-affecting CLI commands call `prompt.ConfirmIO(...)` directly, which means EOF currently surfaces as an error instead of a clean denial. The safer and more ergonomic behavior for those flows is to treat missing confirmation input as “no”.

## What Changes

- Add a shared prompt helper that maps EOF to a clean denial
- Reuse it from config delete, memory clear, graph clear, recovery written-down confirmation, and dead-letter retry
- Add regression coverage for EOF-as-deny behavior at both helper and command levels
- Document the EOF-as-deny contract in specs

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-prompt-helpers`: shared confirmation helpers include an EOF-as-deny wrapper
- `config-cli-commands`: config delete treats EOF as a clean denial
- `cli-agent-memory`: memory clear treats EOF as a clean denial
- `cli-graph-management`: graph clear treats EOF as a clean denial
- `recovery-mnemonic`: written-down confirmation treats EOF as setup abort
- `cli-status-dashboard`: dead-letter retry treats EOF as retry abort through the shared helper

## Impact

- Affected code: `internal/cli/prompt/*`, `internal/cli/configcmd/profile.go`, `internal/cli/memory/clear.go`, `internal/cli/graph/clear.go`, `internal/cli/security/recovery.go`, `internal/cli/status/status.go`
- Affected tests/specs around those commands
- User-facing behavior becomes safer: missing confirmation input no longer raises a hard error in these flows
