## Overview

`lango account session create`, `list`, and `revoke` depend on a live smart account bootstrap and session manager state. Narrow seams around the final create, list, and revoke results are enough to make command-level tests deterministic while keeping production behavior unchanged.

## Decisions

### Introduce small session seams

One seam creates a session and returns the final summary payload. One seam returns the session list payload. One seam performs revocation and returns the final confirmation message. Tests can replace them with deterministic fixtures and avoid bootstrapping the live smart account stack.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Table output uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`
- Empty-state and revoke confirmation use `fmt.Fprintln(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to session semantics
- No change to session payload shapes
