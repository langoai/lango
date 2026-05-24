## Why

The contract CLI family now uses an explicit `--output table|json` contract. If public docs or the main contract interaction spec drift back to a boolean `--output` toggle, operator expectations will break again even if code stays correct.

## What Changes

- add an executable repository guard for stale contract CLI output docs
- document that regression boundary in the docs-only and test-coverage specs

## Impact

- cheaper detection of contract CLI docs drift
- stronger alignment between implemented UX and public operator guidance
