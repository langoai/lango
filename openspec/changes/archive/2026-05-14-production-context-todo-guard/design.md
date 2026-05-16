## Overview

This change broadens a narrow x402-specific placeholder-context guard into a repository-level invariant for production Go code.

## Decision

- Scan only non-test `.go` files under `cmd/` and `internal/`
- Keep test files out of scope so targeted fixture code does not create false positives
- Clean the remaining test-only `context.TODO()` uses anyway so the repository no longer normalizes the pattern in everyday development

## Consequences

- Production code cannot silently regress to placeholder contexts
- Reviewers no longer need to manually spot this class of issue outside x402
