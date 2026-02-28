## Why

Core source files have grown too large as the project scaled (security hardening, settings UI, P2P features). Four files exceeded maintainable size (tools.go 2,709 lines, wiring.go 1,871 lines, forms_impl.go 1,790 lines, types.go 943 lines). Splitting them into domain-focused files within the same package improves navigability and reduces merge conflicts without any API changes.

## What Changes

- Split `internal/app/tools.go` (2,709 lines) into 9 domain-focused files (exec, filesystem, browser, meta, security, automation, p2p, data + orchestrator)
- Split `internal/app/wiring.go` (1,871 lines) into 9 domain-focused files (knowledge, memory, embedding, graph, payment, p2p, automation, librarian + core init)
- Split `internal/cli/settings/forms_impl.go` (1,790 lines) into 6 domain-focused files (knowledge, automation, security, p2p, agent + core forms)
- Split `internal/config/types.go` (943 lines) into 5 domain-focused files (security, knowledge, p2p, automation + root types)
- All splits are same-package file moves — zero API changes, zero import changes for consumers

## Capabilities

### New Capabilities

- `codebase-structure`: File organization conventions and domain-based file splitting rules for the Lango codebase

### Modified Capabilities

_(none — this is a pure structural refactoring with no requirement changes)_

## Impact

- **Code**: 4 large files reorganized into ~24 smaller domain-focused files across 3 packages (`internal/app`, `internal/cli/settings`, `internal/config`)
- **APIs**: No changes — all functions/types remain in the same package with the same signatures
- **Dependencies**: No new dependencies added or removed
- **Build**: `go build ./...` and `go test ./...` pass without changes
- **Risk**: Low — same-package file moves only, no behavioral changes
