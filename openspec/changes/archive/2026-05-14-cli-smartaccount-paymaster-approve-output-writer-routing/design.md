## Overview

`lango account paymaster approve` depends on a live smart account bootstrap, paymaster configuration, and on-chain execute call. A narrow seam around the final approval result is enough to make command-level tests deterministic while keeping production behavior unchanged.

## Decisions

### Introduce a small paymaster-approval seam

The seam returns the fully-shaped approval payload plus a cleanup callback. Tests can replace it with deterministic fixtures and avoid bootstrapping the live smart account stack.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Table output uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to approval semantics
- No change to approval payload shape
