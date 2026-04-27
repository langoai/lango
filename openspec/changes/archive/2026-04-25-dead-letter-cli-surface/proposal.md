## Why

The dead-letter operator surface is currently strongest in the cockpit, but operators also need a non-interactive CLI entrypoint for backlog listing and per-transaction inspection.

## What Changes

- add `lango status dead-letters`
- add `lango status dead-letter <transaction-receipt-id>`
- keep `table` as the default output and support `json`
- reuse the existing dead-letter list/detail read model
- document the landed CLI slice in public docs and main OpenSpec specs

## Impact

- dead-letter triage is available outside the cockpit
- no new backend endpoint or canonical state
- CLI and cockpit now share the same read model
