## Overview

This is a regression-only change. The broker-mode code path already exists and is seam-aware; the goal is to verify that the configured stdin/stdout seams continue to be passed into the broker server on success.

## Decision

- Reuse the existing `runStorageBrokerServerFn`, `mainStdin`, and `mainStdout` seams
- Assert identity of the injected streams rather than reconstructing output through subprocesses

## Consequences

- The top-level broker wrapper contract is fully covered on both success and failure branches
