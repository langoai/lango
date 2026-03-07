## Context

The P2P Economy Layer (45 tasks) was completed and archived. Three automated review agents identified code quality issues. This design covers the 5 highest-value fixes selected from the review findings.

## Goals / Non-Goals

**Goals:**
- Eliminate duplicated `parseUSDC` implementations across budget and risk packages
- Replace stringly-typed action comparisons with typed constants
- Add compile-time interface verification for `noopSettler`
- Prevent potential blocking by moving callback invocation outside mutex lock
- Add capacity hints per Go performance guidelines

**Non-Goals:**
- Changing any public API or behavior
- Refactoring tool handler code structure (intentionally kept flat per review skip rationale)
- Adding Ent persistence for in-memory stores (deferred to future phase)
- Optimizing O(N) scans in CheckExpiry/ListByPeer (MVP scale is sufficient)

## Decisions

### Decision 1: Adapt callers to `wallet.ParseUSDC` signature

`wallet.ParseUSDC` returns `(*big.Int, error)` while the local versions returned `(*big.Int, bool)` and `*big.Int`. Rather than creating a wrapper, each call site adapts directly to the error-returning signature. Budget adds an explicit `Sign() <= 0` check since wallet.ParseUSDC doesn't reject zero values. Risk falls back to default on error (existing behavior preserved).

### Decision 2: String cast for action comparison

`p2pproto.NegotiatePayload.Action` is typed as `string`, while `negotiation.ActionPropose` etc. are `ProposalAction` (alias of `string`). We use `string(negotiation.ActionPropose)` for comparison rather than changing the protocol message type, keeping the P2P protocol layer decoupled from negotiation internals.

### Decision 3: Collect-then-fire pattern for threshold callbacks

Instead of calling `alertCallback` inside the mutex, we collect triggered thresholds into a local slice under lock, then fire callbacks after unlock. This prevents deadlock if the callback (e.g., eventbus.Publish) acquires other locks.

## Risks / Trade-offs

- [Risk] `wallet.ParseUSDC` uses `big.Rat` while old implementations used `big.Float` — minor precision difference possible for edge-case decimal strings → Mitigated: both produce identical results for standard USDC amounts (tested).
- [Risk] Capacity hint `12` in `tools_economy.go` may become stale if tools are added/removed → Acceptable: over-allocation wastes negligible memory, under-allocation just triggers one extra allocation.
