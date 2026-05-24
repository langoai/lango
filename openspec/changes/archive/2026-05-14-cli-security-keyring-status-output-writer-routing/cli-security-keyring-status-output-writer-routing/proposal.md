## Why

`lango security keyring status` still wrote human-readable and JSON output directly to process stdout instead of the Cobra command writer. That left the keyring status surface inconsistent with the already-hardened `security status` command.

## What Changes

- route `security keyring status` text and JSON output through `cmd.OutOrStdout()`
- add command-level writer capture tests using a stub keyring provider
- sync security CLI docs and keyring specs with the output-writer contract

## Impact

- improves automation compatibility and testability for keyring status inspection
- keeps user-visible output unchanged while aligning behavior with the rest of the CLI
