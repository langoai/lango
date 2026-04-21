## Why

The documentation toolchain has moved from MkDocs and Material for MkDocs to a Zensical-native site. This slice records that migration in OpenSpec so the canonical docs contract, the hidden-doc boundary, and the repository-facing docs references all stay aligned.

## What Changes

- Define the canonical documentation site through `zensical.toml`.
- Move hidden docs and withdrawn cockpit sub-guides out of `docs/` so the public site is structurally explicit.
- Update repository docs references to describe the new Zensical toolchain.
- Keep `docs/features/cockpit.md` as the single public cockpit entry after the hidden guide move.

## Impact

- Affected specs: `mkdocs-documentation-site`, `project-docs`, `docs-only`
- Affected docs: `README.md`, `docs/architecture/project-structure.md`, `docs/development/build-test.md`, `docs/features/cockpit.md`
- Affected site behavior: canonical docs build path, public-vs-hidden docs boundary, and cockpit entrypoint clarity
