## Overview

This is a regression-only change for entrypoint ordering.

## Decision

- Reuse the existing `isSandboxWorkerModeFn`, `runSandboxWorkerFn`, `isStorageBrokerModeFn`, and `newRootCmdFn` seams
- Assert that the worker branch returns success and prevents later entrypoint branches from running

## Consequences

- The highest-priority worker-mode branch is explicitly protected from ordering regressions
