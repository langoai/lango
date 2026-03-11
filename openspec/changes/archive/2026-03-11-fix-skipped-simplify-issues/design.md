## Context

A `/simplify` review identified 6 code quality issues. After re-evaluation, 4 warrant fixing (P0–P2), while 2 are intentionally kept (semantic clarity and smart contract abstraction risk). All 4 fixes are localized refactorings with no cross-cutting architectural changes.

## Goals / Non-Goals

**Goals:**
- Eliminate fire-and-forget goroutine that can access a closed store on shutdown
- Reduce DanglingDetector memory footprint by querying only pending escrows
- Consolidate duplicate NoopSettler definitions into a single exported type
- Reduce V1/V2 topic offset code duplication in event monitor

**Non-Goals:**
- Merging `SetDealMappingByDID` methods (semantic clarity preserved)
- Abstracting shared base between HubV2Client/HubClient (smart contract layer risk)
- Changing any external behavior or adding new features

## Decisions

1. **App-level context for goroutine lifecycle**: Add `ctx`/`cancel` fields to `App` struct, created in `New()`, cancelled in `Stop()`. This provides a clean shutdown signal for fire-and-forget goroutines without requiring lifecycle registry registration for simple timer goroutines.
   - Alternative: Pass lifecycle context from `Start()` → rejected because `wireTeamBudgetBridge` is called during `New()`, before `Start()`.

2. **`ListByStatus` on Store interface**: Add a single method rather than a generic query builder. The interface stays minimal and the `DanglingDetector` is the only consumer needing filtered queries.
   - Alternative: Generic `ListWithFilter(predicate)` → over-engineering for one use case.

3. **`escrow.NoopSettler` as exported type**: Place in `internal/economy/escrow/noop_settler.go` alongside the interface it implements. All 3 duplicate definitions (wiring_economy.go, dangling_detector_test.go, bridge tests) replaced.
   - Alternative: Keep per-package test settlers → unnecessary duplication.

4. **Two monitor helpers instead of one**: `extractDealAndAddress` (4 callers) and `extractDealID` (2 callers) kept separate because some events only need dealID. Dispute events excluded from helpers due to different V2 layout (initiator in non-indexed data).

## Risks / Trade-offs

- **Interface change**: Adding `ListByStatus` to `Store` is a breaking change for any external implementors → Mitigated: `Store` is internal-only, and both implementations (memoryStore, EntStore) are updated together.
- **App ctx/cancel scope**: The app-level context is broader than strictly needed for budget bridge goroutines → Acceptable: it's the correct abstraction for "app is shutting down" and can be reused by future goroutines.
