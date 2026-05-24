## Why

The RunLedger CLI docs already described output and limit flags, but the implementation did not actually expose a machine-readable output contract. That left a real UX gap between public guidance and the shipped command surface.

## What Changes

- add `--output table|json` to `lango run list`, `lango run status`, and `lango run journal`
- add practical `--limit` support where the docs already promised it
- reject unknown output values before bootstrap-dependent work and sync docs/specs to the real contract

## Impact

- public docs and shipped CLI behavior now match
- easier machine-readable inspection of RunLedger state
- more consistent operator UX across CLI surfaces
