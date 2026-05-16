## Overview

This change extends the existing top-level stream-discipline work into exit-discipline: entrypoint wrappers may define seam defaults from `os.Exit`, but production paths should not call `os.Exit(...)` directly anymore.

## Decision

- Scan non-test Go files under `cmd/`
- Permit `os.Exit` only on the explicit seam declaration lines already present in `cmd/lango/main.go` and `cmd/zkexport/main.go`
- Treat any other `os.Exit` reference as a regression

## Consequences

- Future wrapper regressions fail fast
- The binary boundary keeps testable exit semantics instead of bypassing seams
