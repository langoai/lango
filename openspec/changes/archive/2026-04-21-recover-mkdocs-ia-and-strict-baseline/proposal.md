## Why

The public documentation site has drifted away from the intended information architecture. Hidden documentation artifacts are still visible to MkDocs, public security deep-dive links are not consistently surfaced, the getting-started quickstart contains a broken installation anchor, and cockpit operator guidance is split across separate public entry points instead of being consolidated on the main cockpit page.

This slice restores a strict MkDocs baseline while keeping the exposed documentation surface intentionally small and truthful.

## What Changes

- Exclude hidden docs and withdrawn operator sub-guides from the MkDocs build.
- Expose only the selected public Security and Research navigation entries.
- Keep the security index aligned with the newly public deep-dive docs.
- Fix the quickstart installation anchor target.
- Consolidate public cockpit operator guidance onto the main cockpit page.

## Impact

- Affected specs: `mkdocs-documentation-site`, `security-docs-sync`, `docs-only`
- Affected docs: `docs/getting-started/quickstart.md`, `docs/features/cockpit.md`, `docs/security/index.md`
- Affected site behavior: MkDocs strict build warnings, public nav exposure, and docs-site coherence
