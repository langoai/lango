# Surface Doctor Bootstrap Errors

## Why

`lango doctor` uses bootstrap to load the encrypted profile and runtime dependencies. When bootstrap fails, the command currently continues with `cfg == nil` and does not preserve the original bootstrap error in the diagnostic results. Operators can see generic configuration failures while the real cause, such as a corrupt envelope or locked profile, disappears from table and JSON output.

## What Changes

- Add a dedicated failing doctor result when bootstrap fails.
- Preserve the original bootstrap error in result details for both table and JSON output.
- Continue running the remaining checks with `cfg == nil` so doctor still provides best-effort environment diagnostics.
- Update CLI docs to describe bootstrap failure reporting.

## Impact

This is a CLI diagnostics correctness change. It does not change bootstrap behavior for normal application startup.
