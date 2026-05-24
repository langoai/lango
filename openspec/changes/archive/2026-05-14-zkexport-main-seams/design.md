## Overview

The core utility behavior already lives in `runZKExport(...)`; the remaining gap is the thin process wrapper in `main()`.

## Decision

- Inject the wrapper's args/stdout/stderr/exit dependencies through package-level seams
- Keep `runZKExport(...)` unchanged
- Test one representative failure path so the wrapper plumbing is locked down

## Consequences

- The binary wrapper becomes regression-testable without subprocesses
- Future refactors cannot silently bypass stderr or exit-code routing
