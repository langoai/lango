## Overview

This guard extends the stream-contract hardening work from CLI command implementations into the top-level `cmd/` entrypoints.

## Decision

- Scan non-test Go files under `cmd/`
- Reject raw `fmt.Print`, `fmt.Printf`, and `fmt.Println`
- Allow direct standard-stream references only on the explicit seam declaration lines already present in `cmd/lango/main.go` and `cmd/zkexport/main.go`

## Consequences

- Future entrypoint stream regressions fail fast
- Wrapper/testability contracts remain stable at the binary boundary
