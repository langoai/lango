## Summary

Remove the production `panic` path from `internal/smartaccount/module` ABI argument initialization while preserving install/uninstall calldata encoding behavior.

## Motivation

The smart account module encoder runs on the account/module management path. Its ABI argument types are deterministic constants, so invalid definitions should surface as returned encoder errors and tests rather than package-init panics.

## Scope

- Replace panic-based ABI type construction with checked initialization that records an error.
- Keep `EncodeInstallModule` and `EncodeUninstallModule` signatures and encoded output unchanged.
- Add a package-level regression guard preventing production `panic` calls in `internal/smartaccount/module`.
- Sync and archive the OpenSpec change.

## Non-Goals

- No changes to module registry behavior.
- No changes to smart account CLI/TUI commands.
- No changes to Safe7579 ABI selectors or calldata layout.
