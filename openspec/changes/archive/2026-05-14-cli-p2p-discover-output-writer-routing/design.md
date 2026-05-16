## Overview

`lango p2p discover` currently builds runtime dependencies, initializes GossipSub discovery, and renders output inside one command body. A narrow data seam allows deterministic command-level tests while preserving production behavior.

## Decisions

### Introduce a discover command data seam

The seam returns:

- the discovered Gossip cards
- a cleanup callback

Tests can stub this seam with fixed card data for empty, table, and JSON paths.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Empty-state text uses `fmt.Fprintln(cmd.OutOrStdout(), ...)`
- Table output uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to discovery matching semantics
- No change to Gossip service behavior outside the new seam
