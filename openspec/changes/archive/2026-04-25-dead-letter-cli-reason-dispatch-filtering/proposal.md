## Why

The dead-letter CLI surface already supports subtype/latest-family and actor/time narrowing, but operators still cannot filter by dead-letter reason text or dispatch reference from the CLI.

## What Changes

- add `--dead-letter-reason-query` to `lango status dead-letters`
- add `--latest-dispatch-reference`
- forward both values unchanged through the existing dead-letter list bridge
- document the landed CLI reason/dispatch-filtering slice in public docs and main OpenSpec specs

## Impact

- better CLI operator triage
- no backend path change
- no detail or retry command change
