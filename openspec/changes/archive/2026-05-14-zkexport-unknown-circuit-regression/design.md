## Overview

This is a regression-only change. The runtime behavior already exists; the goal is to make the actionable failure message explicit and stable.

## Decision

- Reuse the existing deterministic circuit list
- Cover only the top-level unknown-circuit branch in `runZKExport(...)`

## Consequences

- Future edits cannot silently degrade the operator-facing error text for unsupported circuit IDs
