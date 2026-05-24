## Why

The contract CLI family recently normalized `--output` to an explicit `table|json` contract. Without a repository guard, future CLI code could quietly reintroduce boolean `--output` flags and make the operator UX inconsistent again.

## What Changes

- add an executable repository guard that rejects boolean `--output` flag declarations in CLI production code
- document that regression boundary in the main test-coverage spec

## Impact

- stronger consistency for operator-facing output-format UX
- cheaper detection of output-flag regressions during normal test runs
