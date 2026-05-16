## Overview

`lango p2p connect` and `lango p2p disconnect` each perform one runtime operation and then emit a single success line. Small seams around the runtime operations are enough to make command-level output tests deterministic.

## Decisions

### Introduce narrow connect/disconnect seams

The seams return the resolved peer ID string plus a cleanup callback. Tests can stub them without creating a live node.

### Route success output through the Cobra writer

Both commands use `fmt.Fprintf(cmd.OutOrStdout(), ...)` for the final confirmation line.

## Non-Goals

- No change to multiaddr or peer ID parsing behavior
- No change to the underlying P2P network operations
