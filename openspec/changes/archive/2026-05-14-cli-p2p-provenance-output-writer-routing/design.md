## Overview

`lango p2p provenance` is server-backed and mostly delegates to the gateway. For command-level tests, the only unstable dependency is the HTTP JSON POST helper, so a single seam is sufficient.

## Decisions

### Introduce a gateway POST seam

The command group now uses a package-level `provenancePostJSON` seam. Tests can replace it with a deterministic stub and verify both request wiring and output capture without starting a live gateway.

### Route success output through the Cobra writer

Both `push` and `fetch` use `fmt.Fprintf(cmd.OutOrStdout(), ...)` for the final confirmation line.

## Non-Goals

- No change to gateway endpoints or payload shape
- No change to provenance redaction semantics
