## Overview

The helper is unused, so the safest change is deletion rather than further adaptation. The remaining helper functions in `internal/testutil` stay intact.

## Decisions

### Delete instead of deprecate

Because the symbol has no remaining in-repo references, a deprecation cycle inside test-only infrastructure would add noise without benefit.

## Non-Goals

- No runtime behavior changes
- No broader testutil refactor beyond removing the dead helper file
