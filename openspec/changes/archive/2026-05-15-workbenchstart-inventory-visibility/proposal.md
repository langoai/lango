## Why

The bare `lango` workbench relies on `internal/cli/workbenchstart` for context-aware starter and recovery prompts, but that shared support package is currently missing from the public inventory docs.

## What Changes

- add `cli/workbenchstart/` to the architecture inventory with its current responsibilities
- add `workbenchstart/` to the README internal tree
- extend the shared CLI support inventory guard to require that package
- sync the main docs-only and test-coverage specs

## Impact

- more complete public inventory coverage for the bare workbench stack
- better discoverability of prompt-seeding logic
- stronger regression protection against future support-package omissions
