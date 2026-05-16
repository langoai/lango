## Overview

This is a regression-only change around an existing validation rule.

## Decision

- Reuse the existing `isInteractiveFn` seam so mode validation runs in the normal interactive path
- Assert the `unknown mode` message directly from command execution

## Consequences

- Future top-level mode-routing refactors cannot silently weaken mode validation
