## Overview

The top-level `main()` function mixes environment-mode dispatch, command execution, direct stderr writes, and process exit. A small helper boundary is enough to make those failure branches testable without changing runtime behavior.

## Decision

- Add a `runMain()` helper that returns the intended exit code
- Keep worker-mode behavior unchanged
- Route broker-mode and root-command failures through injected stderr
- Let `main()` remain the only place that actually calls the exit seam

## Consequences

- Root failure branches become easy to regression-test
- Direct `os.Exit(1)` no longer bypasses the existing exit seam
