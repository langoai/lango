## Why

The public contract already treated trace-derived metrics as `lango agent trace metrics`, but the actual Cobra wiring exposed them as a root-level `lango agent metrics` subcommand. That made the live CLI diverge from README, CLI index, and main specs.

## What Changes

- move the trace-derived metrics subcommand under `lango agent trace`
- add a regression test so `agent trace metrics` stays nested and root `agent metrics` is not exposed for trace-derived metrics
- sync the main diagnostics spec to the real `trace show` and `trace metrics` command paths

## Impact

- live CLI behavior now matches the published agent diagnostics contract
- operators can use the documented `lango agent trace metrics` path without ambiguity
- stronger regression protection for agent subcommand tree drift
