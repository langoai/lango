# App Boundary Decoupling

## Problem

`internal/cli/status`, `internal/cli/cockpit`, and `internal/cli/agent` directly import `internal/app`. This makes `internal/app` behave like a shared utility package instead of a process composition root and increases the cost of future app decomposition.

## Proposed Change

Make `cmd/lango` the only production importer of `internal/app`. Move app-independent hook registry construction to `internal/toolchain`, remove unused app concrete dependencies from cockpit, inject status dead-letter runtime access from `cmd/lango`, and enforce the boundary with `internal/archtest`.

## User-Facing Impact

Command names, flags, output schemas, status rendering, cockpit status rendering, and `lango agent hooks` output remain compatible.
