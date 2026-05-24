## Why

The first dead-letter CLI surface can already list backlog rows and show per-transaction detail, but operators still cannot narrow the list by the latest retry lifecycle phase from the CLI.

## What Changes

- add `--latest-status-subtype` to `lango status dead-letters`
- add `--latest-status-subtype-family` to `lango status dead-letters`
- validate allowed values explicitly
- forward both values through the existing dead-letter list bridge
- document the landed CLI subtype/latest-family-filtering slice in public docs and main OpenSpec specs

## Impact

- stronger operator triage from the CLI
- no new backend path
- no detail-command change
