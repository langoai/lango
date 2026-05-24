## Why

The docs landing page currently links only a subset of the top-level documentation sections. As more public sections now have their own `index.md`, those sections can become orphaned from the main landing page even though they are intended as first-class navigation entry points.

## What Changes

- add a section catalog to `docs/index.md` that links every top-level docs section carrying its own `index.md`
- add an executable guard that requires `docs/index.md` to keep linking those section indexes
- sync the main docs-only and test-coverage specs

## Impact

- improves top-level discoverability across the public docs
- turns section-level navigation coverage into a maintained invariant
- reduces future landing-page drift as new top-level sections are added
