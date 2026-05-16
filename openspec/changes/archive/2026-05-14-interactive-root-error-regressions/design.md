## Overview

This is a narrow regression-only change around existing top-level TUI startup guardrails.

## Decision

- Reuse the existing `isInteractiveFn` seam
- Assert the returned error text directly from `newRootCmd()` execution

## Consequences

- The non-interactive startup contract for top-level TUI entrypoints is explicit and executable
