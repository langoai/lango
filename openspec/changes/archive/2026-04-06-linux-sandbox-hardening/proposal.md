## Why

`GOOS=linux go build ./...` fails because `isolator_linux.go` calls `NewLandlockIsolator()`/`NewSeccompIsolator()` which have no Linux-side definitions. Additionally, `probePlatform()` uses concrete type-cast (`ll.(*landlockIsolator).abiVersion`) creating tight coupling. Documentation and code comments across 13+ locations falsely claim Linux seccomp/Landlock enforcement is functional when only stubs exist. Users enabling sandbox on Linux get silent no-ops with no indication of the gap.

## What Changes

- Remove concrete type-cast dependency in `probePlatform()` — replace with standalone probe functions
- Add `Reason() string` to `OSIsolator` interface for honest unavailability reporting
- Add `disabledIsolator` type to prevent nil deref when sandbox is disabled by config
- Add `SandboxStatus` struct combining config + runtime state for unified status reporting
- Add `PlatformCapabilities` reason fields (`SeatbeltReason`, `LandlockReason`, `SeccompReason`) with "probe not yet implemented" for unimplemented probes
- Rewrite `isolator_linux.go` to use standalone probe functions and return `noopIsolator` with reason
- Enrich CLI `sandbox status` output with reason-aware formatting and fail-mode explanation
- Fix 13 misleading documentation/code-comment claims about Linux enforcement
- Update TUI settings form descriptions and menu to reflect actual Linux state

## Capabilities

### New Capabilities

- `sandbox-availability-reporting`: Structured availability reporting with reasons — `OSIsolator.Reason()`, `SandboxStatus`, `PlatformCapabilities` reason fields, CLI/TUI honest status output

### Modified Capabilities

- `os-sandbox-core`: `OSIsolator` interface gains `Reason()` method; `PlatformCapabilities` gains reason fields; `probePlatform()` decoupled from concrete types; `noopIsolator` gains reason field; `disabledIsolator` added
- `os-sandbox-cli`: `sandbox status` output restructured with Active Isolation section, reason display, fail-mode explanation, reason-aware capability formatter
- `os-sandbox-integration`: `initOSSandbox()` uses `SandboxStatus` for logging; reason included in warn messages

## Impact

- **Interface change**: `OSIsolator` gains `Reason()` — all implementations and mocks must update (3 test mocks, 5 stub/real implementations, 1 composite)
- **Build fix**: `GOOS=linux go build ./...` transitions from broken to passing
- **Documentation**: README, 3 docs pages, 1 config reference, 4 code comment blocks, 4 TUI strings corrected
- **No behavioral change on macOS**: Seatbelt isolation continues to work identically
