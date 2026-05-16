## Why

The codebase now routes startup notices for bare `lango`, `lango cockpit`, and `lango chat` through seam-aware stderr paths, but the public core CLI docs only mention that behavior for `lango chat`. That leaves the top-level docs behind the actual entrypoint contracts.

## What Changes

- Update `docs/cli/core.md` to mention startup stderr seam routing for bare `lango` workbench and `lango cockpit`
- Record the docs truth-sync expectation in the docs-only spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `docs-only`: top-level startup stream docs stay aligned with current entrypoint behavior

## Impact

- Affected docs: `docs/cli/core.md`
- Affected specs: `openspec/specs/docs-only/spec.md`
- No code or runtime behavior changes
