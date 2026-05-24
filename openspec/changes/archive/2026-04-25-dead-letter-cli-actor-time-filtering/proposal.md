## Why

The dead-letter CLI surface already supports query, adjudication, subtype, latest-family filtering, and the first retry action, but operators still cannot narrow the backlog by latest manual replay actor or latest dead-letter time window from the CLI.

## What Changes

- add `--manual-replay-actor` to `lango status dead-letters`
- add `--dead-lettered-after`
- add `--dead-lettered-before`
- validate both time flags as RFC3339
- forward the three values through the existing dead-letter list bridge
- document the landed CLI actor/time-filtering slice in public docs and main OpenSpec specs

## Impact

- stronger CLI triage for operators
- no backend path change
- no detail or retry command change
