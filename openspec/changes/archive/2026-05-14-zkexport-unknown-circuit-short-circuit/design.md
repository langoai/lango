## Overview

The single-circuit command path already has all information needed to reject an unknown circuit before touching gnark setup.

## Decision

- Move unknown-circuit validation ahead of prover service construction in the single-circuit path
- Keep `--all` behavior unchanged because it necessarily needs the prover service

## Consequences

- Unknown circuit failures stay focused on the real operator mistake
- Broken prover-service setup no longer masks invalid-circuit input
