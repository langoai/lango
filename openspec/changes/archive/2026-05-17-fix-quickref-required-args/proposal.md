# Fix Quick Reference Required Args

## Why

The README and CLI index quick references abbreviate several provenance and P2P
commands so much that copy-pasted commands fail immediately with Cobra argument
or required-flag errors. Dedicated CLI pages already document the correct forms,
but the high-traffic quick references are stale.

## What Changes

- Update provenance quick references with required labels, session keys, bundle
  files, and `--run` where required.
- Update the P2P reputation quick reference with required `--peer-did <did>`.
- Add docs guard coverage so quick references cannot regress to non-runnable
  bare command forms.

## Impact

- Modified capabilities: `downstream-docs-sync`, `docs-only`.
- No runtime behavior changes.
- Public docs become safer for copy-paste setup and operator usage.
