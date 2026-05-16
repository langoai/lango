## Overview

This change turns a recently cleaned spec-hygiene issue into an executable repository invariant.

## Decision

- Scan only main specs under `openspec/specs`
- Reject the specific archive-generated placeholder strings that should never survive in main specs
- Keep the guard narrow and deterministic so it stays cheap to run in the normal Go test suite

## Consequences

- Future archive placeholder regressions fail fast
- Reviewers no longer need to catch this specific issue manually every time
