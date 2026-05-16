## Overview

`lango p2p status` currently builds runtime dependencies and renders output inside a single command body. A small data seam keeps runtime behavior intact while letting tests replace the live node with deterministic fixtures.

## Decisions

### Introduce a status command data seam

The command resolves a small `statusCommandData` struct before rendering:

- peer ID
- listen addresses
- connected peer count
- max peers
- mDNS status
- relay status
- ZK handshake status

The seam also returns a cleanup callback so production behavior stays unchanged.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Text output uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`

## Non-Goals

- No change to status payload fields
- No change to P2P node initialization behavior beyond the test seam
