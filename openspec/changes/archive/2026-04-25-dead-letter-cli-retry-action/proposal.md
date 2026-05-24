## Why

The dead-letter CLI surface can already list backlog rows and inspect per-transaction detail, but operators still need a first non-cockpit recovery action to request a retry from the CLI.

## What Changes

- add `lango status dead-letter retry <transaction-receipt-id>`
- precheck retryability through the existing detail status surface
- require confirmation by default
- support `--yes` to skip the prompt
- reuse the existing `retry_post_adjudication_execution` control path
- document the landed CLI retry-action slice in public docs and main OpenSpec specs

## Impact

- first CLI recovery action for dead-lettered executions
- no new backend write path
- CLI stays aligned with the same control plane as the cockpit
