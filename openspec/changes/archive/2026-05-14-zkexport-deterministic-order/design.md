## Overview

The command already has a single helper for enumerating circuit IDs. Sorting there is enough to stabilize both usage text and all-mode export iteration.

## Decision

- Sort the output of `circuitIDs()`
- Drive `--all` iteration from that sorted list instead of ranging the circuit map directly

## Consequences

- Wrapper and CI output become stable across runs
- Tests can assert concrete ordering without flakiness
