## Why

Top-level interactive entrypoints currently defer `--mode` validation too far into their startup flow, and the `health` utility emits duplicate Cobra-managed error text on failures. Both behaviors make CLI failures noisier and less predictable than they need to be.

## What Changes

- validate interactive top-level `--mode` values immediately after bootstrap config load and before TUI startup/app construction
- silence duplicate Cobra error emission for `lango health` while preserving returned errors
- add regression coverage for the preflight validation path
- sync CLI docs and main specs

## Impact

- invalid top-level modes fail earlier with fewer side effects
- `lango health` failure runs no longer print duplicated error text
- wrapper and automation behavior stays cleaner and more deterministic
