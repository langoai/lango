## Overview

`lango p2p pricing` is a pure config-backed inspection command. It does not require runtime seams because tests can exercise the command directly with a synthetic bootstrap result.

## Decisions

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Overview and tool-specific text output use `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`

### Cover both overview and tool-specific views

Tests cover:

- the full overview text output
- the single-tool text output
- the JSON output

## Non-Goals

- No change to pricing config shape
- No change to pricing selection semantics
