## Overview

`lango p2p firewall` is mostly config-backed inspection and guidance output. No runtime seam is required; tests can exercise the commands directly with synthetic config.

## Decisions

### Route all non-error output through the Cobra writer

- `list --json` uses `json.NewEncoder(cmd.OutOrStdout())`
- `list` table uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`
- `add` and `remove` guidance text use `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`

### Cover all operator-facing views

Tests cover:

- list empty-state
- list table
- list JSON
- add guidance
- remove guidance

## Non-Goals

- No change to firewall rule config shape
- No change to runtime firewall semantics
