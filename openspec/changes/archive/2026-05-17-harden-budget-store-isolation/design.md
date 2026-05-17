# Design

## Root Cause

`Store` holds `*TaskBudget` values and currently returns the same pointers to callers. `TaskBudget` also contains mutable nested values: `*big.Int` fields and `[]SpendEntry`, where each entry has an `Amount *big.Int`. Copying only the struct would still leak nested mutable state.

## Approach

- Introduce a store-local clone helper for `TaskBudget`.
- Clone `TotalBudget`, `Spent`, `Reserved`, and each `SpendEntry.Amount`.
- Clone the entries slice even when it is empty, preserving nil versus empty behavior only where it is not semantically relevant.
- Store internal copies and return copies from public store methods.

## Non-Goals

- Do not change the `Guard` interface.
- Do not introduce persistence or external storage.
- Do not change public CLI output or documentation.

## Downstream Impact

Budget engine methods already call `Store.Update` after mutating budget snapshots, so they should keep working. Tests that directly mutate returned store values must update the store explicitly to persist changes.
