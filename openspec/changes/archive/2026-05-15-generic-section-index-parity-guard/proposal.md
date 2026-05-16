## Why

The repository now has multiple section-specific landing-page guards, but the underlying invariant is broader: every public section index should link all of the dedicated pages in its own directory. Encoding that rule once makes future docs growth safer and reduces the chance of orphaned pages in any section.

## What Changes

- add a generalized executable guard that validates every `docs/*/index.md` against sibling `*.md` pages
- sync the main docs-only and test-coverage specs

## Impact

- turns section-level docs completeness into a reusable invariant
- reduces future orphan-page drift across all public docs sections
- complements the narrower home, CLI, architecture, and features guards
