## Why

`lango payment balance` still writes both table and JSON output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves this payment inspection surface inconsistent with the already-hardened `payment x402` command.

## What Changes

- route `payment balance` table output through `cmd.OutOrStdout()`
- route `payment balance --json` output through the same writer
- add command-level coverage that confirms no direct output bypass happens before dependency errors
- sync payment CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for payment balance inspection
- keeps user-visible output unchanged while aligning with the rest of the CLI
