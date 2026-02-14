## Why

Observational Memory, Secret Management, and Output Scanning are implemented at the Core layer but have no CLI/TUI exposure. Users cannot manage memory entries, store/delete secrets, view security status, or diagnose these subsystems from the command line. The onboard wizard also lacks configuration for Observational Memory, leaving users to edit JSON manually.

## What Changes

- Add `lango memory list|status|clear` commands to manage observational memory entries per session
- Add `lango security secrets list|set|delete` commands to manage encrypted secrets via CLI
- Add `lango security status` command to display security configuration overview
- Add `ObservationalMemoryCheck` to `lango doctor` for validating OM configuration
- Add `OutputScanningCheck` to `lango doctor` for verifying interceptor/secret alignment
- Add Observational Memory configuration form to the onboard TUI wizard
- Add `DeleteReflectionsBySession` to the memory store (required by `memory clear`)

## Capabilities

### New Capabilities

- `cli-memory-management`: CLI commands for listing, inspecting status, and clearing observational memory entries per session
- `cli-secrets-management`: CLI commands for storing, listing, and deleting encrypted secrets with local crypto provider
- `cli-security-status`: CLI command to display aggregated security configuration and state

### Modified Capabilities

- `cli-doctor`: Add ObservationalMemoryCheck (validates thresholds, budget consistency, provider existence) and OutputScanningCheck (validates interceptor/secret alignment)
- `cli-onboard`: Add Observational Memory form to the onboard wizard menu with fields for enabled, provider, model, and token threshold configuration
- `observational-memory`: Add `DeleteReflectionsBySession` store method to support bulk session cleanup

## Impact

- **New packages**: `internal/cli/memory/` (4 files)
- **New files**: `internal/cli/security/crypto_init.go`, `secrets.go`, `status.go`; `internal/cli/doctor/checks/observational_memory.go`, `output_scanning.go`
- **Modified files**: `internal/memory/store.go`, `store_test.go`, `cmd/lango/main.go`, `internal/cli/security/migrate.go`, `internal/cli/doctor/checks/checks.go`, `internal/cli/onboard/forms_impl.go`, `menu.go`, `wizard.go`, `state_update.go`
- **Dependencies**: Uses existing `session.EntStore`, `memory.Store`, `security.SecretsStore`, `security.KeyRegistry`, `security.LocalCryptoProvider`
- **No breaking changes**: All additions are new commands/checks/forms; existing behavior unaffected
