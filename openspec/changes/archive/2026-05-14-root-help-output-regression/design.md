## Overview

This is a narrow regression-only change. The code path already exists; the goal is to make its output-routing contract explicit and executable.

## Decision

- Reuse the existing `isInteractiveFn` seam
- Assert against the command output buffer directly
- Keep the contract in `cli-reference`, alongside the other top-level entrypoint stream guarantees
