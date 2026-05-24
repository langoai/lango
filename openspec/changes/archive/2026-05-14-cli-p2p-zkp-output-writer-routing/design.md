## Overview

`lango p2p zkp` is a pure inspection surface: one config-backed command and one static-data command. No runtime seam is required; tests can execute the commands directly with synthetic config.

## Decisions

### Route all non-error output through the Cobra writer

- `status --json` uses `json.NewEncoder(cmd.OutOrStdout())`
- `status` text uses `fmt.Fprintln` / `fmt.Fprintf` against `cmd.OutOrStdout()`
- `circuits --json` uses `json.NewEncoder(cmd.OutOrStdout())`
- `circuits` table uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`

### Align docs to actual runtime output

The public docs previously described fields like compiled circuit count and compilation status that the current command does not print. This change aligns those examples to the real output shape.

## Non-Goals

- No change to ZKP config shape
- No change to the set of available circuits
