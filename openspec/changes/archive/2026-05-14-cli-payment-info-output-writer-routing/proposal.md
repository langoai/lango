## Why

`lango payment info` still writes both table and JSON output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves this payment inspection surface inconsistent with the already-hardened `payment x402` and `payment balance` commands.

## What Changes

- route `payment info` table output through `cmd.OutOrStdout()`
- route `payment info --json` output through the same writer
- add a lightweight payment-deps seam and command-level capture tests
- sync payment CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for payment info inspection
- keeps user-visible output unchanged while aligning with other commands
