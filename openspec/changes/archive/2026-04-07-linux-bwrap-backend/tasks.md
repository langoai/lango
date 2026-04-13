## 1. Stage 1 — Policy → bwrap argv compiler

- [x] 1.1 Add `internal/sandbox/os/bwrap_args.go` with platform-agnostic `compileBwrapArgs(Policy) ([]string, error)`
- [x] 1.2 Map `ReadOnlyGlobal`, `ReadPaths`, `WritePaths`, `DenyPaths`, and `Network` modes to bwrap flags
- [x] 1.3 Reuse existing `sanitizePath()` for injection safety on every path
- [x] 1.4 Reject `DenyPaths` that are missing or non-directory (PR-level constraint, file deny deferred)
- [x] 1.5 Add `internal/sandbox/os/bwrap_args_test.go` with table-driven tests for default/strict/MCP policies, NetworkUnixOnly, NetworkAllow, ReadPaths-not-global, injection 4 cases, deny path file/missing, and empty policy
- [x] 1.6 Verify `go test ./internal/sandbox/os/ -run TestCompileBwrapArgs` passes
- [x] 1.7 Verify `go build ./...` and `GOOS=linux GOARCH=amd64 go build ./...` are clean

## 2. Stage 2 — BwrapIsolator (build-tag split)

- [x] 2.1 Add `internal/sandbox/os/bwrap_linux.go` (`//go:build linux`) with `BwrapIsolator` struct and `NewBwrapIsolator()` factory
- [x] 2.2 Probe via `exec.LookPath("bwrap")` + `filepath.Abs` and store the resolved absolute path on the struct
- [x] 2.3 Capture `bwrap --version` output as the `version` field
- [x] 2.4 Implement `Apply()` that calls `compileBwrapArgs`, sets `cmd.Path = b.resolvedPath`, and writes `cmd.Args[0]` to the same absolute path (not the bare string `"bwrap"`)
- [x] 2.5 Insert `"--"` separator and append the original argv after the bwrap flags
- [x] 2.6 Add `Available()`, `Name()`, `Reason()`, and `Version()` methods
- [x] 2.7 Add `internal/sandbox/os/bwrap_other.go` (`//go:build !linux`) with a `NewBwrapIsolator()` stub returning `Reason()="bwrap is Linux-only"`
- [x] 2.8 Remove `bwrapStub` type and `NewBwrapStub()` from `internal/sandbox/os/registry.go`
- [x] 2.9 Update `PlatformBackendCandidates()` darwin/linux branches to call `NewBwrapIsolator()` instead of `NewBwrapStub()`
- [x] 2.10 Add `internal/sandbox/os/bwrap_linux_test.go` (`//go:build linux`) with host-dependent probe assertions, Apply-wraps-command verification (resolvedPath + `"--"` separator), and unavailable / available-reason cases
- [x] 2.11 Update `internal/sandbox/os/registry_test.go`: replace `TestBwrapStub` with `TestNewBwrapIsolator_NameAndInterface` (cross-platform contract); switch `TestSelectBackend_AutoAggregatesCandidateReasons` and `TestSelectBackend_ExplicitPreservesIdentity` to `fakeIsolator` so they no longer depend on host bwrap state
- [x] 2.12 Verify build, tests, and lint stay clean on both native and `GOOS=linux` cross-build

## 3. Stage 3 — Real Linux kernel probes

- [x] 3.1 Replace `probeLandlockKernel()` stub in `internal/sandbox/os/isolator_linux.go` with `unix.Syscall(SYS_LANDLOCK_CREATE_RULESET, 0, 0, LANDLOCK_CREATE_RULESET_VERSION)` that returns `(true, abi, "Landlock ABI N")` on success and `(false, 0, "Landlock not supported by this kernel ...")` on ENOSYS
- [x] 3.2 Replace `probeSeccompKernel()` stub with `unix.PrctlRetInt(PR_GET_SECCOMP, ...)` returning a reason that explicitly states the result is a presence signal only and does NOT prove BPF filter capability
- [x] 3.3 Add `readProcSelfSeccompMode()` helper that augments the seccomp reason with `/proc/self/status:Seccomp` when readable
- [x] 3.4 Update `noopIsolator` reason to point users to `backend=bwrap` instead of "not yet implemented"
- [x] 3.5 Update `internal/sandbox/os/probe.go` doc comments: `HasSeccomp` carries the "presence signal only" caveat; `LandlockReason` example includes `"Landlock ABI 3"`
- [x] 3.6 Update `Summary()` to drop the dead `"unknown (probe not yet implemented)"` branch and replace it with an honest `"linux (no Landlock or seccomp interface detected)"` fallback
- [x] 3.7 Update `internal/cli/sandbox/sandbox.go` `capabilityReasonStatus` doc comment example to match the new reason texts (the function body keeps its defensive `"not yet implemented"` branch)
- [x] 3.8 Run `go mod tidy` to promote `golang.org/x/sys` from indirect to direct
- [x] 3.9 Verify build, cross-build, tests, and lint are all clean

## 4. Stage 4 — Smoke test expansion and reliability fix

- [x] 4.1 Add hidden `lango sandbox _probe-net <addr>` cobra subcommand (`Hidden: true`, `Args: cobra.ExactArgs(1)`) that calls `net.DialTimeout("tcp", addr, 2*time.Second)` and exits non-zero on connect failure
- [x] 4.2 Register the hidden subcommand in `NewSandboxCmd`
- [x] 4.3 Add `versioner` optional interface (`interface { Version() string }`) for backends that capture a version string
- [x] 4.4 Rewrite `newTestCmd` body to use a results slice + loop with four cases (write restriction, read permission, workspace write, network deny) and print the optional `Version()` line
- [x] 4.5 Add `discardOutput(*exec.Cmd)` helper that sets both `Stdout` and `Stderr` to `io.Discard`
- [x] 4.6 Convert `runWriteTest` to invoke `/usr/bin/touch` directly (no shell) with `discardOutput`
- [x] 4.7 Convert `runReadTest` to invoke `/bin/cat` directly (no shell) with `discardOutput`
- [x] 4.8 Add `runWorkspaceWriteTest` that creates a temp dir via `os.MkdirTemp`, resolves it through `filepath.EvalSymlinks` for macOS realpath matching, adds it to a policy's `WritePaths`, and touches a file inside via `/usr/bin/touch` + `discardOutput`
- [x] 4.9 Add `runNetworkDenyTest` that opens an ephemeral `127.0.0.1:0` listener, accepts in a goroutine, re-invokes `os.Executable() sandbox _probe-net <addr>` as a sandboxed child, and returns true only when `c.Run() != nil`
- [x] 4.10 Verify on macOS Seatbelt with `lango sandbox test`: 4/4 PASS
- [x] 4.11 Verify build, cross-build, tests, and lint stay clean

## 5. Stage 5 — Downstream documentation sync

- [x] 5.1 Update `internal/sandbox/os/isolator.go` package doc comment: replace "Linux: not yet implemented (planned)" with the bwrap-on-Linux description; mark `NewOSIsolator()` as a backwards-compat helper and recommend `ParseBackendMode + SelectBackend`
- [x] 5.2 Update `internal/config/types_sandbox.go`: split `Backend` doc into bwrap (requires bubblewrap) vs native (planned); rewrite `NetworkMode` doc with bwrap `--unshare-net` mapping; clarify `AllowedNetworkIPs` is ignored on bwrap; mark `OSSandboxConfig.SeccompProfile` as NOT YET ENFORCED and consumed by the planned native backend (bwrap ignores it)
- [x] 5.3 Update `internal/cli/settings/forms_sandbox.go`: enabled, backend, network mode, and seccomp profile field descriptions
- [x] 5.4 Update `README.md`: feature list "OS-level Sandbox" line, sandbox config table `sandbox.backend` row, and `sandbox.os.seccompProfile` row
- [x] 5.5 Update `docs/cli/sandbox.md`: bwrap requirement note, `lango sandbox test` four-case walkthrough, hidden `_probe-net` mention, `io.Discard` rationale
- [x] 5.6 Update `docs/configuration.md`: `sandbox.backend`, `sandbox.networkMode`, `sandbox.allowedNetworkIPs`, and `sandbox.os.seccompProfile` rows
- [x] 5.7 Verify build and full test suite stay clean

## 6. Stage 6 — OpenSpec change (this artifact set)

- [x] 6.1 Run `openspec new change linux-bwrap-backend`
- [x] 6.2 Write `proposal.md` with Why, What Changes, Capabilities, and Impact
- [x] 6.3 Write `design.md` with Context, Goals/Non-Goals, Decisions D1–D10, Risks/Trade-offs, Migration Plan, and Open Questions
- [x] 6.4 Write `specs/linux-bwrap-isolation/spec.md` (NEW capability) with `BwrapIsolator`, resolved path, version capture, non-Linux stub, `compileBwrapArgs`, and directory-only deny requirements
- [x] 6.5 Write `specs/os-sandbox-core/spec.md` MODIFIED `Platform capability detection` requirement covering real Linux probes and the `HasSeccomp` caveat
- [x] 6.6 Write `specs/sandbox-backend-registry/spec.md` MODIFIED `PlatformBackendCandidates` requirement, REMOVED `Stub isolators for planned backends`, and ADDED `Native stub remains a planned backend`
- [x] 6.7 Write `specs/os-sandbox-cli/spec.md` MODIFIED `sandbox test command honors configured backend` requirement (four cases + io.Discard contract) and ADDED `sandbox workspace write smoke test` and `sandbox network deny smoke test via hidden self-subcommand`
- [x] 6.8 Write `tasks.md` (this file) with all stage tasks marked done
- [x] 6.9 Run `/opsx:verify linux-bwrap-backend`
- [x] 6.10 Sync delta specs into `openspec/specs/`
- [x] 6.11 Run `/opsx:archive linux-bwrap-backend`
