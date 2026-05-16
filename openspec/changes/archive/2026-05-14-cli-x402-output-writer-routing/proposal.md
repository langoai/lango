## Why

`lango payment x402` still writes both table and JSON output directly to process stdout instead of the Cobra command writer. That breaks output capture for wrappers and tests and leaves this payment inspection surface inconsistent with the other hardened CLI commands.

## What Changes

- route `payment x402` table output through `cmd.OutOrStdout()`
- route `payment x402 --json` output through the same writer
- add command-level capture tests for both modes
- sync X402 CLI specs and docs with the output-writer contract

## Impact

- improves testability and automation compatibility for X402 inspection
- keeps user-visible output unchanged while aligning with the rest of the CLI
