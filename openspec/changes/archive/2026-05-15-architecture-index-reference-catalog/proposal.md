## Why

The architecture landing page currently links all dedicated architecture references, but that coverage is not protected by an executable guard. New architecture pages could be added later and silently become orphaned from the top-level architecture index.

## What Changes

- add an executable guard that requires every `docs/architecture/*.md` page other than `index.md` to stay linked from `docs/architecture/index.md`
- sync the main docs-only and test-coverage specs

## Impact

- turns architecture-reference discoverability into a maintained invariant
- reduces future deep-dive architecture doc drift
- catches orphaned architecture pages early
