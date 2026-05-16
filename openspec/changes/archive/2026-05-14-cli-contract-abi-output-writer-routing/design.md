## Overview

`lango contract abi load` is a local file parsing command, so tests can execute it directly with temporary ABI fixtures. No runtime seam is required.

## Decisions

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Text summary uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`

### Cover both output modes directly

Tests use small temporary ABI files and assert:

- text summary capture
- JSON payload capture

## Non-Goals

- No change to ABI cache behavior
- No change to parsing or count semantics
