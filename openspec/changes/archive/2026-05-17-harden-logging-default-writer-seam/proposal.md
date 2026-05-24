# Proposal: Harden Logging Default Writer Seam

## Why

Default logs currently write to process stdout when no explicit writer or output path is configured. stdout is the CLI result channel, so default diagnostic output can contaminate command output and is hard to verify without process-global stream interception.

## What Changes

- Route the default logging output path through a package-level writer seam.
- Default that seam to stderr to keep stdout reserved for user-facing command output.
- Preserve explicit `LogConfig.Writer` and `LogConfig.OutputPath` precedence.
- Add a focused regression test proving default logging writes through the seam.

## Impact

Runtime logging remains opt-in through `logging.Init`. Callers that need stdout logs can still pass `LogConfig.Writer`. File output behavior is unchanged.
