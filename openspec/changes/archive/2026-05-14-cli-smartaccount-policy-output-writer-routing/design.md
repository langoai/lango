## Overview

`lango account policy show` and `lango account policy set` depend on a live smart account bootstrap, account lookup, and policy engine state. Narrow seams around the final show and update results are enough to make command-level tests deterministic while keeping production behavior unchanged.

## Decisions

### Introduce small policy seams

One seam returns the fully-shaped policy summary payload. Another seam applies updates and returns the final update summary payload. Tests can replace them with deterministic fixtures and avoid bootstrapping the live smart account stack.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Table output uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to policy semantics
- No change to policy payload shapes
