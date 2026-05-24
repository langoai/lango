## Why

`lango security kms status` still wrote human-readable and JSON output directly to process stdout instead of the Cobra command writer. That left the KMS status surface inconsistent with the hardened `security status` and `keyring status` commands.

## What Changes

- route `kms status` text output through `cmd.OutOrStdout()`
- route `kms status --json` output through the same writer
- add command-level writer capture tests using config-backed bootloaders
- sync security CLI docs and cloud-KMS spec with the output-writer contract

## Impact

- improves automation compatibility and testability for KMS status inspection
- keeps user-visible output unchanged while aligning behavior with the rest of the CLI
