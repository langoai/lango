## Why

The standalone mission workbench now has a richer first-run contract than the first-touch CLI docs describe. Runtime and detailed feature docs already explain the readiness split: incomplete profiles point to `lango onboard`, `lango settings`, and `lango doctor`, while ready profiles expose starter prompts and the `Enter` / `1-3` quick-start path. The CLI index and quickstart guide still describe bare `lango` only as a generic workbench launch.

## What Changes

- Update the CLI reference index to describe the readiness-aware bare-`lango` entry flow.
- Update the getting-started quickstart tip so new operators see the setup-recovery and starter-prompt split immediately.
- Record the sync in OpenSpec.

## Impact

- Improves first-touch operator guidance without changing runtime behavior.
- Brings public entry-point docs in line with the actual workbench UX.
