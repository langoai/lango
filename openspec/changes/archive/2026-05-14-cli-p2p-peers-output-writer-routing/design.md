## Overview

`lango p2p peers` constructs peer info and renders it inside the command body. A narrow data seam lets tests replace the live runtime with deterministic peer fixtures while keeping production behavior unchanged.

## Decisions

### Introduce a peers command data seam

The seam returns a slice of peer info records plus a cleanup callback. This is enough to validate empty-state, table, and JSON rendering without booting a live node in tests.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Empty-state text uses `fmt.Fprintln(cmd.OutOrStdout(), ...)`
- Table output uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to peer payload fields
- No change to live P2P connection behavior
