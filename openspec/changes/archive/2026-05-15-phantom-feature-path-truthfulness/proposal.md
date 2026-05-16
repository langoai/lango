## Why

The `phantom-feature-wiring` main spec still named a deleted companion discovery file and implied that the old discovery slice was the relevant source of truth. That stale path and wording make the spec less trustworthy than the current runtime model.

## What Changes

- sync the `phantom-feature-wiring` main spec to the current gateway-backed companion model
- extend the broken-path guard so the deleted companion discovery path cannot silently return

## Impact

- better alignment between the main spec and the current runtime
- less confusion about companion connectivity versus removed discovery code
- stronger regression protection for another deleted path
