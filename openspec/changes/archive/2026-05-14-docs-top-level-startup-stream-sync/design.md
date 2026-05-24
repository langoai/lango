## Overview

This is a documentation truth-sync cleanup. The runtime and tests already enforce the startup stderr seam behavior; the public docs just need to catch up.

## Decision

- Update only the public CLI core reference
- Add a small docs-only requirement so future top-level startup stream changes stay documented

## Consequences

- Public CLI docs remain consistent with the current top-level entrypoint contracts
