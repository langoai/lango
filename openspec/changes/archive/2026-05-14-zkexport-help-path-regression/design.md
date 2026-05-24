## Overview

This is a regression-only change around existing `flag.ErrHelp` handling.

## Decision

- Reuse the existing prover-service seam to prove help short-circuits before setup
- Assert the exit code and stderr help text directly from `runZKExport(...)`

## Consequences

- Future refactors cannot accidentally turn `--help` into a failing or heavyweight path
