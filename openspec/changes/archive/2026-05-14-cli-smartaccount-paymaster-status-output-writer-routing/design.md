## Overview

`lango account paymaster status` depends on a live paymaster provider behind smart account bootstrap. A narrow seam around the final status result is enough to make command-level tests deterministic while keeping production behavior unchanged.

## Decisions

### Introduce a small paymaster-status seam

The seam returns the fully-shaped paymaster status payload plus a cleanup callback. Tests can replace it with deterministic fixtures and avoid bootstrapping the live smart account stack.

### Route all non-error output through the Cobra writer

- JSON uses `json.NewEncoder(cmd.OutOrStdout())`
- Table output uses `tabwriter.NewWriter(cmd.OutOrStdout(), ...)`

## Non-Goals

- No change to paymaster provider semantics
- No change to paymaster status payload shape
