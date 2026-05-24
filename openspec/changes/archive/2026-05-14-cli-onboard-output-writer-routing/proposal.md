## Why

`lango onboard` still writes preset banners, cancel messages, and post-save guidance directly to process stdout. That makes wrapper capture and command-level testing inconsistent with the rest of the CLI writer-routing hardening work.

## What Changes

- Route onboard preset and completion output through `cmd.OutOrStdout()`
- Make the next-steps printer writer-aware
- Update tests, docs, and OpenSpec with the writer-routing contract

## Impact

- Makes onboard completion output consistent with the CLI writer-routing hardening work
- Removes process-global stdout swapping from onboard tests
