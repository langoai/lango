# Harden Budget Store Isolation

## Why

The in-memory budget store currently exposes internal `TaskBudget` pointers through `Allocate`, `Get`, and `List`, and stores caller-owned pointers in `Update`. Because `TaskBudget` contains mutable `*big.Int` fields and a mutable entries slice, callers can mutate budget state without going through `Store.Update`. That bypasses `UpdatedAt`, weakens auditability, and can undermine spending enforcement invariants.

## What Changes

- Return detached `TaskBudget` snapshots from `Allocate`, `Get`, and `List`.
- Store detached snapshots on `Update` so later caller mutations cannot alter persisted budget state.
- Deep-copy mutable budget fields, including `*big.Int` values and spend entry amounts.
- Add focused regression tests for allocation, get, list, and update isolation.

## Impact

This is an internal core hardening change. Callers that intentionally modify a budget must continue to call `Store.Update`; mutating returned snapshots alone no longer changes stored state.
