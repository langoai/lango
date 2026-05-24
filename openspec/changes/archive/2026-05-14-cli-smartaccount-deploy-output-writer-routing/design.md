## Overview

`lango account deploy` depends on a live smart account bootstrap and deployment lookup. A narrow seam around the final deployment result is enough to make command-level tests deterministic while keeping production behavior unchanged.

## Decisions

### Introduce a small deploy-result seam

The seam returns the fully-shaped deployment payload plus a cleanup callback. Tests can replace it with deterministic fixtures and avoid bootstrapping the live smart account stack.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Table output uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to deployment semantics
- No change to deployment payload shape
