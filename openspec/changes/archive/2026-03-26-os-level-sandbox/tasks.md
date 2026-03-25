## 1. OS Sandbox Core Package

- [x] 1.1 Create `internal/sandbox/os/errors.go` with ErrIsolatorUnavailable, ErrSandboxRequired, ErrInvalidPolicy
- [x] 1.2 Create `internal/sandbox/os/policy.go` with Policy, FilesystemPolicy, NetworkPolicy, ProcessPolicy types and DefaultToolPolicy/StrictToolPolicy/MCPServerPolicy presets
- [x] 1.3 Create `internal/sandbox/os/isolator.go` with OSIsolator interface and noopIsolator
- [x] 1.4 Create `internal/sandbox/os/seatbelt_profile.go` with GenerateSeatbeltProfile via text/template and path sanitization
- [x] 1.5 Create `internal/sandbox/os/seatbelt_darwin.go` with SeatbeltIsolator (sandbox-exec wrapping)
- [x] 1.6 Create `internal/sandbox/os/seatbelt_stub.go` for non-darwin
- [x] 1.7 Create `internal/sandbox/os/landlock_stub.go` and `seccomp_stub.go` for non-linux
- [x] 1.8 Create `internal/sandbox/os/probe.go` with PlatformCapabilities and Probe()
- [x] 1.9 Create `internal/sandbox/os/composite.go` for multi-isolator chaining
- [x] 1.10 Create build-tag platform files: isolator_darwin.go, isolator_linux.go, isolator_other.go
- [x] 1.11 Create `internal/sandbox/os/policy_test.go` with table-driven tests for policy, profile generation, path sanitization, IP validation, probe

## 2. Config

- [x] 2.1 Create `internal/config/types_sandbox.go` with SandboxConfig and OSSandboxConfig
- [x] 2.2 Add Sandbox field to Config struct in types.go
- [x] 2.3 Add sandbox defaults in loader.go DefaultConfig()

## 3. Exec Tool Integration

- [x] 3.1 Add OSIsolator, SandboxPolicy, FailClosed fields to exec.Config
- [x] 3.2 Implement applySandbox() method on exec.Tool
- [x] 3.3 Insert applySandbox() in Run() before cmd.Run()
- [x] 3.4 Insert applySandbox() in RunWithPTY() before pty.Start()
- [x] 3.5 Insert applySandbox() in StartBackground() before cmd.Start()
- [x] 3.6 Add CleanupProfileFile() calls after process completion
- [x] 3.7 Add exec sandbox tests (nil isolator, available, fail-open, fail-closed)

## 4. MCP Transport Integration

- [x] 4.1 Add isolator field to ServerConnection
- [x] 4.2 Add SetOSIsolator() method
- [x] 4.3 Apply isolator in createTransport() stdio branch with MCPServerPolicy
- [x] 4.4 Add MCP sandbox tests (with/without isolator, error non-fatal, non-stdio unaffected)

## 5. Skill Script Integration

- [x] 5.1 Add isolator and workspacePath fields to skill.Executor
- [x] 5.2 Add SetOSIsolator() method
- [x] 5.3 Apply isolator in executeScript() with DefaultToolPolicy(workspacePath)
- [x] 5.4 Add CleanupProfileFile() call after script run
- [x] 5.5 Add skill sandbox tests

## 6. App Wiring

- [x] 6.1 Create internal/app/wiring_sandbox.go with initOSSandbox() and sandboxPolicy()

## 7. CLI Commands

- [x] 7.1 Create internal/cli/sandbox/sandbox.go with NewSandboxCmd
- [x] 7.2 Implement `lango sandbox status` subcommand
- [x] 7.3 Implement `lango sandbox test` subcommand
- [x] 7.4 Register sandbox command in cmd/lango/main.go under "sys" group

## 8. Verification

- [x] 8.1 go build ./... passes
- [x] 8.2 go test ./internal/sandbox/os/ passes
- [x] 8.3 go test ./internal/tools/exec/ passes
- [x] 8.4 go test ./internal/mcp/ passes
- [x] 8.5 go test ./internal/skill/ passes
- [x] 8.6 go test ./internal/config/ passes
