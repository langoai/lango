## Overview

`lango p2p reputation` is a storage-backed inspection command. A narrow data seam around the reputation lookup is enough to make command-level output tests deterministic while keeping runtime behavior unchanged.

## Decisions

### Introduce a reputation lookup seam

The seam returns the `reputation.PeerDetails` record for a peer DID. Tests can replace it with deterministic fixtures for:

- missing-record output
- text rendering
- JSON rendering

### Remove dead logger setup

The command created a logger that was never used and then checked an impossible nil branch. This change removes that dead path while keeping user-facing behavior intact.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Missing-record and text output use `fmt.Fprintf` / `fmt.Fprintln` against `cmd.OutOrStdout()`

## Non-Goals

- No change to reputation scoring semantics
- No change to storage payload shape
