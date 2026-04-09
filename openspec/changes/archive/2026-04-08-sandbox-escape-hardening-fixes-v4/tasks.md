## 1. bwrap two-phase smoke probe (D1)

- [x] 1.1 Extend `BwrapIsolator` struct with `networkIsolation bool` + `networkIsolationReason string` fields
- [x] 1.2 Rewrite `NewBwrapIsolator` to run `smokeProbeBwrapBase` (fatal) then `smokeProbeBwrapNetwork` (partial-degradation)
- [x] 1.3 Add `smokeProbeBwrapBase(abs string) error` and `smokeProbeBwrapNetwork(abs string) error` — both reuse `compileBwrapArgs` for argv, 2-second timeout each
- [x] 1.4 Add Apply-time network gate: reject `NetworkDeny`/`NetworkUnixOnly` before `compileBwrapArgs` call when `networkIsolation==false`, preserve `cmd.Path`/`cmd.Args` on reject
- [x] 1.5 Add `NetworkIsolationAvailable() bool` and `NetworkIsolationReason() string` exported methods; `Reason()` stays empty when `Available()==true`
- [x] 1.6 Add unit tests: `TestNewBwrapIsolator_NetworkIsolationContract_HostDependent`, `TestBwrapIsolator_ApplyRejectsNetworkDenyWhenDowngraded`, `TestBwrapIsolator_ApplyPermitsNetworkAllowWhenDowngraded`, `TestSmokeProbeBwrapBase_HostDependent`, `TestSmokeProbeBwrapNetwork_HostDependent`
- [x] 1.7 CLI status: add `networkIsolator` interface + partial-degradation line in `internal/cli/sandbox/sandbox.go` Active Isolation section
- [x] 1.8 Verify Stage 1: `go build ./...`, `GOOS=linux GOARCH=amd64 go build ./...`, `GOOS=linux go vet ./internal/sandbox/os/...`, `go test -count=1 ./...`, `golangci-lint run ./...`

## 2. `.git` walk-up discovery (D2)

- [x] 2.1 Add private `findGitRoot(workDir string) string` helper to `internal/sandbox/os/policy.go`: walk from `filepath.Abs(workDir)` to parent via `filepath.Dir`, terminate when `parent==cur`, return first ancestor `.git` that `isDir` accepts
- [x] 2.2 Update `DefaultToolPolicy` to use `findGitRoot(workDir)` instead of `filepath.Join(workDir, ".git")` for baseline deny
- [x] 2.3 Add `TestFindGitRoot` subtests: direct parent, 2-level ancestor, deeply nested, worktree file skip, empty input, filesystem root termination
- [x] 2.4 Add `TestDefaultToolPolicy_WalksUpToGitRoot` regression guard: nested subdir → ancestor `.git` appears in DenyPaths
- [x] 2.5 Verify Stage 2: same full verification recipe as Stage 1

## 3. `MCPServerPolicy` workspace signature + wiring (D3)

- [x] 3.1 Grep `MCPServerPolicy` across entire repo to confirm call-site inventory
- [x] 3.2 Extend `MCPServerPolicy(dataRoot)` to `MCPServerPolicy(workDir, dataRoot)`; reuse `findGitRoot(workDir)` for baseline deny
- [x] 3.3 Update `internal/sandbox/os/policy_test.go` existing `TestMCPServerPolicy*` cases for 2-arg signature; add `TestMCPServerPolicy_DenyWorkspaceGit` + `TestMCPServerPolicy_WorkspaceGitPlusDataRoot`
- [x] 3.4 Update `internal/sandbox/os/bwrap_args_test.go:TestCompileBwrapArgs_MCPServerPolicy` for 2-arg signature
- [x] 3.5 Add `workspacePath string` field to `ServerConnection` in `internal/mcp/connection.go`
- [x] 3.6 Change `ServerConnection.SetOSIsolator(iso, dataRoot)` → `SetOSIsolator(iso, workspacePath, dataRoot)`
- [x] 3.7 In `createTransport`, set `cmd.Dir = sc.workspacePath` when non-empty; call `MCPServerPolicy(sc.workspacePath, sc.dataRoot)`
- [x] 3.8 Add `workspacePath string` field to `ServerManager` in `internal/mcp/manager.go`
- [x] 3.9 Change `ServerManager.SetOSIsolator(iso, dataRoot)` → `SetOSIsolator(iso, workspacePath, dataRoot)`; propagate in connection creation loop
- [x] 3.10 Update `internal/mcp/connection_test.go` two `SetOSIsolator` call sites for 3-arg signature
- [x] 3.11 Update `internal/app/wiring_mcp.go` to resolve `cfg.Sandbox.WorkspacePath` with `os.Getwd()` fallback and pass it to `mgr.SetOSIsolator`
- [x] 3.12 Verify Stage 3: same full verification recipe

## 4. Documentation sync + OpenSpec archive (D4 + meta)

- [x] 4.1 Update `docs/cli/sandbox.md` with two-phase smoke probe mention and Network Iso partial degradation UX line
- [x] 4.2 Create OpenSpec change `sandbox-escape-hardening-fixes-v4` via `openspec new change`
- [x] 4.3 Write `proposal.md` covering D1+D2+D3+CLI surfacing
- [x] 4.4 Write `design.md` covering D1-D4 rationale + worktree trade-off explicit note
- [x] 4.5 Write this `tasks.md`
- [x] 4.6 Write delta specs for 4 capability specs: `linux-bwrap-isolation`, `os-sandbox-core`, `os-sandbox-integration`, `mcp-integration`
- [ ] 4.7 Run `openspec validate sandbox-escape-hardening-fixes-v4 --strict`
- [ ] 4.8 Run `openspec archive sandbox-escape-hardening-fixes-v4 -y` (or `--no-validate` with commit message if main spec `## Purpose` header is still missing, per PR 5b deferral)
- [ ] 4.9 Stop at commit boundary — user commits manually
