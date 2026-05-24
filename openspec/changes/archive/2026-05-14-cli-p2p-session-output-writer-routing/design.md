## Overview

`lango p2p session` mixes ephemeral node creation, session-store operations, and output rendering in the command bodies. Narrow seams around list/revoke operations are enough to make command-level tests deterministic while preserving production behavior.

## Decisions

### Introduce small seams for list and revoke operations

The command group now uses three small seams:

- active session list lookup
- single-session revoke
- revoke-all

Tests can replace these seams with deterministic fixtures without booting a live node.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Empty-state, table output, and revoke confirmations use `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`
- Table output uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to handshake session semantics
- No change to invalidation reasons or session payload shape
