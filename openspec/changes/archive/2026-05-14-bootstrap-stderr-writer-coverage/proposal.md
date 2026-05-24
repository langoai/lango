## Why

Bootstrap stderr behavior is already specified for secure-storage offer outcomes, but additional stderr paths like envelope migration and keyfile shredding were still tied to global stderr without deterministic regression coverage.

## What Changes

- Route the remaining bootstrap stderr writes through the existing `bootstrapErrWriter` seam
- Add deterministic bootstrap coverage for envelope migration banner and keyfile shred warning

## Impact

- Completes stderr seam coverage for bootstrap’s operator-facing warnings
- Makes bootstrap diagnostics safer to refactor without changing runtime behavior
