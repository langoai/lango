## Why

The repository ships many dedicated CLI reference pages under `docs/cli/`, but the top-level CLI index remains mostly a quick-reference surface. Without an explicit catalog, deeper command-family docs become harder to discover and future dedicated pages can drift out of the index.

## What Changes

- add a dedicated-reference catalog section to `docs/cli/index.md`
- add an executable guard that requires every `docs/cli/*.md` page other than `index.md` to stay linked from the index
- sync the main docs-only and test-coverage specs

## Impact

- improves discoverability of deeper CLI docs from the top-level index
- turns dedicated-page coverage into a maintained invariant
- reduces future CLI reference drift
