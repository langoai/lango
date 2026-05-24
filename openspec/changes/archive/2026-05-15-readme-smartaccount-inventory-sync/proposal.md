## Why

The README internal CLI inventory still uses old hyphen-compressed shorthand for the smart-account family, while the architecture inventory and dedicated smart-account docs already describe the current slash-separated command surface.

## What Changes

- update the README internal tree smart-account row to the current slash-separated command surface
- update the smart-account inventory guard so README is validated against the same contract as the architecture inventory
- sync the main docs-only and test-coverage specs with the current smart-account inventory wording

## Impact

- consistent smart-account inventory wording across public docs
- less ambiguity about the actual command paths
- stronger regression protection against stale README shorthand
