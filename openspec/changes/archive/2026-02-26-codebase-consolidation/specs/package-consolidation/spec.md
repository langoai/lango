# Spec: Package Consolidation

## Overview
Merge three underused packages into their logical parent packages to improve codebase clarity.

## Requirements

### R1: ctxutil → types
- Move `Detach()` function and `detachedCtx` type from `internal/ctxutil/` to `internal/types/context.go`
- Move tests to `internal/types/context_test.go`
- Update all importers to use new path
- Delete `internal/ctxutil/` directory

#### Scenarios
- **Background task**: `types.DetachContext(ctx)` preserves `Value()` but detaches from cancellation.
- **No import cycle**: `types` package has no upstream dependencies.

### R2: passphrase → security/passphrase
- Move all files from `internal/passphrase/` to `internal/security/passphrase/`
- Package name remains `passphrase`
- Update all importers (bootstrap.go, bootstrap_test.go)
- Delete `internal/passphrase/` directory

#### Scenarios
- **Passphrase acquisition**: Priority order (keyring → keyfile → interactive → stdin) unchanged.
- **Keyfile operations**: Read/Write/Shred/ValidatePermissions unchanged.

### R3: zkp → p2p/zkp
- Move all files from `internal/zkp/` to `internal/p2p/zkp/` (including `circuits/` subdirectory)
- Package names remain `zkp` and `circuits`
- Update all importers (wiring.go, internal cross-references)
- Delete `internal/zkp/` directory

#### Scenarios
- **ZKP proving/verifying**: `ProverService` functionality unchanged.
- **Circuit compilation**: All 4 circuits (ownership, attestation, capability, balance) work identically.

## Constraints
- Zero functional changes — only import paths change
- No import cycles introduced
- All existing tests must pass without modification
