## Why

`lango status` and its dead-letter subcommands document `--output` as accepting only `table` or `json`, but the implementation does not validate that contract. Invalid values like `--output yaml` currently fall through silently and can trigger unnecessary bootstrap or bridge work before eventually rendering table output.

## What Changes

- validate and normalize `--output` values across all `lango status` command paths
- reject unknown output formats before bootstrap or dead-letter bridge loading
- add regression coverage for root and dead-letter paths
- sync README, CLI docs, and status specs with the enforced contract

## Impact

- more predictable operator UX for `lango status`
- less accidental backend work on invalid invocations
- tighter alignment between implementation, docs, and OpenSpec
