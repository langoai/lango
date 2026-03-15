## Why

Full audit of the tool subsystem initialization revealed **systemic bugs** repeated across 12+ subsystems: missing config validation, absent lifecycle registration, silent failures without warning logs, and disabled categories not registered with the tool catalog. These cause agent tool discovery failures, goroutine leaks on shutdown, and `builtin_health` diagnostics that cannot report disabled subsystems.

## What Changes

- Add `Validate()` method to `SmartAccountConfig` for required field pre-validation (A1)
- Integrate config validation in both app wiring and CLI paths (A1, A6)
- Add `Stop()` method to `SessionGuard` and register it with lifecycle registry (A3)
- Add warning logs for risk engine skip, sentinel skip, X402 secrets nil, and observability sub-flag conflicts (A2, A4, C1, F1)
- Add RPCURL pre-validation in payment wiring to prevent confusing `ethclient.Dial("")` errors (B1)
- Register disabled categories for all 15+ subsystems so `builtin_health` can report them (E1, A5, B2, D1)
- Update smart account disabled description to list required config fields (A5)

## Capabilities

### New Capabilities

- `tool-discovery-diagnostics`: Config validation, warning logs, and disabled category registration for tool subsystem initialization stability

### Modified Capabilities

- `smart-account-init-validation`: Add config validation (Validate method) and lifecycle registration for SessionGuard
- `tool-catalog`: Register disabled categories for all subsystems (15+ missing registrations)

## Impact

- `internal/config/types_smartaccount.go` — new Validate() method
- `internal/app/wiring_smartaccount.go` — validation, lifecycle param, warning logs
- `internal/app/wiring_payment.go` — RPCURL pre-validation, X402 warning
- `internal/app/wiring_observability.go` — sub-flag conflict warning
- `internal/app/app.go` — registry param, 15+ disabled category registrations
- `internal/economy/escrow/sentinel/session_guard.go` — Stop() method + active flag guard
- `internal/cli/smartaccount/deps.go` — Validate() integration
- New test files for Validate() and Stop()
