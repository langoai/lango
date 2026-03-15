## Why

After migrating the bundler client to EntryPoint v0.7 format and adding CirclePermitProvider, downstream artifacts (CLI, TUI, docs, tests) still referenced the v0.6 EntryPoint address and lacked the new `mode` config field. This change synchronizes all downstream artifacts with the core changes.

## What Changes

- Replace all 27 occurrences of v0.6 EntryPoint address (`0x5FF137D4b0FDCD49DcA30c7CF57E578a026d2789`) with v0.7 (`0x0000000071727De22E5E9d8BAf0edAc6f37da032`) across tests and docs
- Add `Mode` field to CLI `paymaster status` output (table and JSON)
- Add permit mode support to CLI `deps.go` `initPaymasterProvider()`
- Add `sa_paymaster_mode` TUI form field and state update handler
- Update documentation: features, CLI reference, configuration table

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `paymaster`: CLI/TUI/docs updated to expose and document the `mode` field
- `smart-account`: All v0.6 EntryPoint address references replaced with v0.7

## Impact

- **Tests**: 7 test files updated (wallet, smartaccount, bundler, paymaster)
- **Docs**: 3 documentation files updated (features, cli, configuration)
- **CLI**: 2 files updated (paymaster.go, deps.go)
- **TUI**: 2 files updated (forms_smartaccount.go, state_update.go)
- No behavioral changes — purely address migration and UI/doc sync
