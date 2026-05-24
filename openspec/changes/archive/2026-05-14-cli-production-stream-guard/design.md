## Overview

This change codifies the current CLI stream-routing discipline as an executable repository invariant.

## Decision

- Scan non-test Go files under `internal/cli`
- Reject raw `fmt.Print`, `fmt.Printf`, and `fmt.Println`
- Reject direct `os.Stdout` and `os.Stderr` references except in the small set of explicit seam files that intentionally define default writers

## Consequences

- CLI command implementations stay capturable through Cobra streams
- Future drift toward process-global output becomes a fast test failure instead of a review-only concern
