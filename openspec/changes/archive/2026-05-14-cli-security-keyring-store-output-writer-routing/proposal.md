## Why

`lango security keyring store` still wrote non-error status messages directly to process stdout instead of the Cobra command writer. That left its success and already-stored flows inconsistent with the hardened security CLI surfaces.

## What Changes

- route `keyring store` already-stored output through `cmd.OutOrStdout()`
- route `keyring store` success output through the same writer
- add command-level writer capture coverage for the already-stored path
- sync security CLI docs and keyring specs with the output-writer contract

## Impact

- improves automation compatibility and testability for keyring storage status messages
- keeps user-visible output unchanged while aligning behavior with the rest of the CLI
