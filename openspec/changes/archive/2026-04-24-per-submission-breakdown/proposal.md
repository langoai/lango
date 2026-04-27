## Why

Dead-letter backlog rows already expose transaction-global retry and family summaries, but operators still cannot see which submission contributed what across a transaction's history.

## What Changes

- expose compact `submission_breakdown` on dead-letter backlog rows
- describe the new row field in public docs
- sync the main OpenSpec docs-only and meta-tools specs

## Impact

- better operator visibility into cross-submission retry history
- no canonical state changes
- no new tools or stores
