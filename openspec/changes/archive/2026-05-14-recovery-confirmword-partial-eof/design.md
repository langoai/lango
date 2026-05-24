## Overview

The shared raw line reader intentionally returns partial content together with EOF so callers can decide whether the content is still meaningful. Recovery confirmation-word prompts should treat that case as valid when the supplied word already matches.

## Decision

- Preserve hard read failures for non-EOF errors
- Preserve failure for EOF with empty content
- Accept EOF with non-empty content when the normalized word matches the expected mnemonic word

## Consequences

- Recovery setup behaves more consistently with the underlying shared line-reader contract
- Seam-driven and wrapper-driven executions no longer need an artificial trailing newline for success
