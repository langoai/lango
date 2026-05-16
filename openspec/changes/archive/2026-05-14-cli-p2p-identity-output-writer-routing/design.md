## Overview

`lango p2p identity` currently mixes runtime dependency creation and output rendering inside the command body. For reliable command-level tests, this change introduces a narrow seam that resolves identity command data before rendering.

## Decisions

### Introduce an identity command data seam

A package-level loader returns:

- DID
- peer ID
- key storage mode
- listen addresses
- cleanup callback

This keeps the runtime behavior unchanged while letting tests replace the loader with deterministic fixtures.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Text output uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`

## Non-Goals

- No changes to DID resolution semantics
- No changes to the P2P runtime bootstrap path outside the new test seam
