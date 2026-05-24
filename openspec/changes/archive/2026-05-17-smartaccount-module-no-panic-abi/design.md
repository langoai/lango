## Context

`internal/smartaccount/module/abi_encoder.go` defines the shared ABI argument list for ERC-7579 module install/uninstall calldata. The current helper panics if a hard-coded ABI type fails to parse.

## Design

Create ABI argument definitions through a helper that returns `(abi.Arguments, error)`. Store both the arguments and initialization error at package scope. `EncodeInstallModule` and `EncodeUninstallModule` will check the initialization error before packing and return an actionable encoding error instead of relying on a package-init panic path.

This preserves public API signatures and existing successful encoding output.

## Testing

Add an AST-based package guard that fails if non-test Go files in `internal/smartaccount/module` call `panic`. Existing encoder tests continue to verify selectors, layout, module type, address encoding, and byte payload behavior.

## Risks

The main risk is accidentally changing ABI argument order or selector bytes. Existing calldata layout tests cover those invariants, so this change only alters initialization error handling.
