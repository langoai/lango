## Overview

This change protects a narrow but repeatedly corrected docs drift: stale shared confirmation examples that omit the colon separator before the entered answer.

## Decision

- Scan only public docs under `docs/` plus `README.md`
- Reject the concrete stale example shape `[y/N] y|n|yes|no`
- Keep the guard intentionally narrow so prose references to `[y/N]` prompts remain valid

## Consequences

- Future drift in public examples fails fast
- The guard stays low-noise and cheap to maintain
