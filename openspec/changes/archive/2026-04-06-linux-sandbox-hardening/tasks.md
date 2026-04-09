## 1. Interface & Core Types

- [x] 1.1 Add `Reason() string` to `OSIsolator` interface in `isolator.go`
- [x] 1.2 Add `reason string` field to `noopIsolator` and implement `Reason()`
- [x] 1.3 Add `disabledIsolator` type with `Reason() = "sandbox disabled by configuration"`

## 2. PlatformCapabilities & SandboxStatus

- [x] 2.1 Add `SeatbeltReason`, `LandlockReason`, `SeccompReason` fields to `PlatformCapabilities` in `probe.go`
- [x] 2.2 Update `Summary()` to return `"unknown (...)"` for Linux unimplemented probes
- [x] 2.3 Create `status.go` with `SandboxStatus` struct and `NewSandboxStatus()` constructor

## 3. Linux Build Fix & Probe Decoupling

- [x] 3.1 Rewrite `isolator_linux.go`: remove `NewLandlockIsolator`/`NewSeccompIsolator` calls, return `noopIsolator` with reason
- [x] 3.2 Implement standalone `probeLandlockKernel()`, `probeSeccompKernel()`, `linuxKernelVersion()` in `isolator_linux.go`
- [x] 3.3 Set `LandlockReason`/`SeccompReason` to `"probe not yet implemented"` in `probePlatform()`

## 4. Existing Implementation Reason() Methods

- [x] 4.1 Add `Reason()` to `SeatbeltIsolator` in `seatbelt_darwin.go`
- [x] 4.2 Add `Reason()` to `SeatbeltIsolator` stub in `seatbelt_stub.go`
- [x] 4.3 Add `Reason()` to `compositeIsolator` in `composite.go`
- [x] 4.4 Add `Reason()` to `landlockIsolator` stub in `landlock_stub.go`
- [x] 4.5 Add `Reason()` to `seccompIsolator` stub in `seccomp_stub.go`
- [x] 4.6 Update `isolator_darwin.go` to pass reason to noop fallback and populate `SeatbeltReason`
- [x] 4.7 Update `isolator_other.go` to pass reason to noop and populate all reason fields

## 5. Consumer Updates

- [x] 5.1 Add `Reason()` to `mockIsolator` in `exec_test.go`
- [x] 5.2 Add `Reason()` to `mockIsolator` in `executor_test.go` (skill)
- [x] 5.3 Add `Reason()` to `mockIsolator` in `connection_test.go` (mcp)
- [x] 5.4 Update `wiring_sandbox.go` to use `SandboxStatus` and log `Reason()` + `Summary()`
- [x] 5.5 Replace `capabilityStatus()` with `capabilityReasonStatus()` in `sandbox.go` CLI
- [x] 5.6 Restructure `sandbox status` output: add Active Isolation section, fail-mode explanation
- [x] 5.7 Update TUI form descriptions in `forms_sandbox.go` (3 fields)
- [x] 5.8 Update menu description in `menu.go`

## 6. Documentation Corrections

- [x] 6.1 Fix package doc in `errors.go`
- [x] 6.2 Fix interface/function docs in `isolator.go`
- [x] 6.3 Fix config field comments in `types_sandbox.go`
- [x] 6.4 Fix `SetOSIsolator` comment in `skill/executor.go`
- [x] 6.5 Fix `README.md` sandbox feature description and config table
- [x] 6.6 Fix `docs/index.md` feature description
- [x] 6.7 Fix `docs/features/index.md` feature description
- [x] 6.8 Fix `docs/cli/sandbox.md` note block
- [x] 6.9 Fix `docs/configuration.md` config reference table

## 7. Codex Review Fixes

- [x] 7.1 Fix `exec.go` `applySandbox()`: enforce `FailClosed` even when `OSIsolator` is nil
- [x] 7.2 Fix `sandbox.go` CLI: pass nil isolator when `sandbox.enabled=false` so `disabledIsolator` is used
- [x] 7.3 Fix `capabilityReasonStatus()`: only "probe not yet implemented" → `unknown`, definitive negatives → `unavailable (reason)`
- [x] 7.4 Add fail-closed warning logs to `wiring_knowledge.go` for skill executor path
- [x] 7.5 Add fail-closed warning logs to `wiring_mcp.go` for MCP manager path
- [x] 7.6 Fix `sandbox.go` CLI: hide fail-mode wording when `sandbox.enabled=false`

## 8. Verification

- [x] 8.1 `go build ./...` (macOS)
- [x] 8.2 `GOOS=linux GOARCH=amd64 go build ./...` (Linux cross-compile)
- [x] 8.3 `go test ./...` (all tests pass, including updated mocks)
- [x] 8.4 `golangci-lint run ./...` (0 issues)
