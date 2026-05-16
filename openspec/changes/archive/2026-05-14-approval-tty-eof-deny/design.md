## Overview

The TTY approval fallback already uses seam-injected streams and the shared raw line reader. The remaining inconsistency is how it handles EOF: it currently bubbles the read error instead of returning a denial.

## Decision

- Keep the existing non-terminal failure mode unchanged
- Keep explicit `y`/`yes` and `a`/`always` semantics unchanged
- Map `io.EOF` from the shared raw line reader to a normal denial response with provider metadata
- Preserve hard errors for non-EOF read failures

## Consequences

- Terminal approval fallback aligns with the safer confirmation behavior already used in CLI destructive commands
- Tests can assert the denial path deterministically with injected streams
- Operators who close stdin or send EOF during approval do not get a noisy read failure
