## Overview

Bootstrap already had a partial stderr seam from recent secure-storage work. Two remaining operator-facing messages still bypassed it: the legacy envelope migration banner and the keyfile shred warning. This change simply folds those into the same seam and verifies them.

## Decisions

### Reuse `bootstrapErrWriter`

No new abstraction is needed. The existing seam is extended to cover the remaining bootstrap stderr writes so tests can capture them uniformly.

### Cover the two remaining operator-facing branches

Tests exercise:
- legacy envelope migration banner emission
- keyfile shred warning emission

## Non-Goals

- No change to migration semantics
- No change to keyfile shredding behavior
