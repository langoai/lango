## Why

Several `lango metrics` subcommands still write table, JSON, or empty-state output directly to process stdout. That makes wrapper capture and command-level testing awkward.

## What Changes

- Route `lango metrics sessions`, `tools`, `agents`, and `history` output through `cmd.OutOrStdout()`
- Make the shared JSON/tabwriter helpers writer-aware
- Add command-level regression coverage using `httptest` gateways
- Update docs and OpenSpec with the writer-routing contract

## Impact

- Makes the remaining metrics breakdown commands consistent with the CLI writer-routing hardening work
- Improves testability without changing endpoint payload semantics
