## Why

RunLedger Phase 1 has the basic journal/snapshot/PEV skeleton in place, but P1-level gaps remain in actual execution authority and workspace lifecycle. Specifically, proposal journaling without step ownership verification and branch naming that is not retry-safe directly violate Task OS core invariants. This change brings Phase 1 to a safe hardened state and creates a stable foundation for subsequent Phase 2-4 OpenSpec changes to build upon.

## What Changes

- Modify `run_propose_step_result` to verify step existence, owner agent, and allowed state before journal append.
- Make workspace preparation retry-safe so that branch/path lifecycle does not break on retries and repeated verification.
- Make it explicit in code and documentation that RunLedger module intentionally keeps workspace isolation disabled in Phase 1.
- Clean up RunLedger README/docs/OpenSpec to match current actual behavior.
- Plan Phase 2-4 as separate OpenSpec changes so subsequent implementation can proceed immediately.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `run-ledger`: step proposal authorization, retry-safe workspace lifecycle, and
  explicit phase-gated workspace activation semantics

## Impact

- `internal/runledger/tools.go`
- `internal/runledger/workspace.go`
- `internal/app/modules_runledger.go`
- `internal/runledger/tools_test.go`
- `internal/runledger/workspace_test.go` (new)
- `README.md`
- `docs/features/run-ledger.md`
- `openspec/specs/run-ledger/spec.md`
