## Why

The features landing page presents many capability cards, but some dedicated feature pages such as `agent-format.md`, `learning.md`, and `zkp.md` are not directly linked from the index. That makes those deep dives easier to orphan as the feature set grows.

## What Changes

- add a feature-reference catalog to `docs/features/index.md`
- add an executable guard that requires `docs/features/index.md` to keep linking every dedicated feature page
- sync the main docs-only and test-coverage specs

## Impact

- improves discoverability of all feature deep dives from the features landing page
- turns feature-page coverage into a maintained invariant
- reduces future features-index drift
