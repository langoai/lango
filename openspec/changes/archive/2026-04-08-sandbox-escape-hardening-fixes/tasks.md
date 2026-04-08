## 1. Seatbelt template read-deny (Stage 1)
- [x] 1.1 Modify `internal/sandbox/os/seatbelt_profile.go` template so each `DenyPaths` entry emits both `(deny file-read* (subpath "{{.}}"))` and `(deny file-write* (subpath "{{.}}"))`
- [x] 1.2 Extend `TestGenerateSeatbeltProfile` table in `internal/sandbox/os/policy_test.go` with two cases using manually-constructed `Policy{DenyPaths: [...]}`: single DenyPath asserts both read and write deny lines; multiple DenyPaths asserts each entry produces both lines
- [x] 1.3 Verify build (`go build ./... && GOOS=linux GOARCH=amd64 go build ./...`), tests (`go test ./...`), lint (`golangci-lint run ./internal/sandbox/os/...`)

## 2. `.git` + dataRoot isDir guard (Stage 2)
- [x] 2.1 Add `isDir(p string) bool` private helper in `internal/sandbox/os/policy.go` and import `"os"`
- [x] 2.2 Gate `.git` baseline deny in `DefaultToolPolicy` on `isDir(gitPath)`; drop any mention of unconditional `.git` addition from the doc comment
- [x] 2.3 Gate `dataRoot` deny in both `DefaultToolPolicy` and `MCPServerPolicy` on `isDir(absDataRoot)`
- [x] 2.4 Rewrite `TestDefaultToolPolicy` and `TestDefaultToolPolicy_EmptyDataRoot` in `internal/sandbox/os/policy_test.go` to use `t.TempDir()` + `os.Mkdir(filepath.Join(workDir, ".git"), 0o755)`
- [x] 2.5 Rewrite `TestStrictToolPolicy` to use `t.TempDir()` with a real `.git` directory
- [x] 2.6 Rewrite `TestMCPServerPolicy` to use `t.TempDir()` as dataRoot
- [x] 2.7 Replace `TestGenerateSeatbeltProfile` table cases that depended on `DefaultToolPolicy(literal, literal)` with directly-constructed `Policy{}` entries (the string-literal paths no longer stat as directories)
- [x] 2.8 Add `TestDefaultToolPolicy_MissingGitNotDenied` (no `.git`, expect empty DenyPaths)
- [x] 2.9 Add `TestDefaultToolPolicy_GitFileNotDenied` (worktree: `.git` as file, expect empty DenyPaths)
- [x] 2.10 Add `TestDefaultToolPolicy_MissingDataRootNotDenied`
- [x] 2.11 Add `TestMCPServerPolicy_MissingDataRoot`
- [x] 2.12 Add `TestCompileBwrapArgs_DefaultToolPolicy_NoGitDir` in `internal/sandbox/os/bwrap_args_test.go` using `t.TempDir()` without `.git` — verify `compileBwrapArgs` succeeds
- [x] 2.13 Add `TestCompileBwrapArgs_DefaultToolPolicy_GitFile` using `t.TempDir()` with `.git` as a regular file — verify `compileBwrapArgs` succeeds
- [x] 2.14 Verify build cross-platform, tests, lint

## 3. Skill executor publish restore (Stage 3)
- [x] 3.1 Delete the early return at `isolator == nil && failClosed` in `internal/skill/executor.go` `executeScript`
- [x] 3.2 Move the nil-isolator decision (rejected + skipped branches) above temp file creation so the publish path runs before any `os.CreateTemp` allocation
- [x] 3.3 Remove the now-dead `else if e.failClosed` / `else` branches after `isolator.Apply` block
- [x] 3.4 Add `sync` and `eventbus` imports to `internal/skill/executor_test.go`
- [x] 3.5 Add `TestExecuteScript_FailClosedWithoutIsolatorPublishesRejection` with a real `eventbus.Bus` and `SubscribeTyped[SandboxDecisionEvent]` collector, asserting the event is published before `ErrSandboxRequired` is returned
- [x] 3.6 Add `TestExecuteScript_FailOpenWithoutIsolatorPublishesSkipped` verifying the fail-open nil-isolator path publishes `"skipped"` and still runs the script
- [x] 3.7 Verify existing `TestExecuteScript_FailClosed_NilIsolator`, `TestExecuteScript_FailClosed_ApplyError`, `TestExecute_Script_WithIsolator` still pass
- [x] 3.8 Verify build, tests, lint

## 4. sandbox status single bootstrap (Stage 4)
- [x] 4.1 Change `newStatusCmd` signature in `internal/cli/sandbox/sandbox.go` to take only `BootLoader` (remove cfgLoader parameter)
- [x] 4.2 Derive `cfg` from `boot.Config` inside status RunE; `defer boot.DBClient.Close()` to match the existing cobra convention
- [x] 4.3 Update `NewSandboxCmd` internal wiring: `newStatusCmd(bootLoader)` while keeping `newTestCmd(cfgLoader)` (external signature preserved)
- [x] 4.4 Change `renderRecentDecisions` signature to take `*bootstrap.Result` directly instead of a `BootLoader` callback; the helper no longer re-invokes the loader
- [x] 4.5 Pass the already-resolved `boot` from status RunE to `renderRecentDecisions`
- [x] 4.6 Update `sandbox_test.go`: remove `errors` import, rename `TestRenderRecentDecisions_NilBootLoaderSilent` → `TestRenderRecentDecisions_NilBootSilent`, delete `TestRenderRecentDecisions_BootLoaderErrorSilent` (bootloader error is now handled at the outer status level), update `TestRenderRecentDecisions_NilDBClientSilent` to pass `&bootstrap.Result{}` directly
- [x] 4.7 Verify `cmd/lango/main.go` does NOT need changes (external `NewSandboxCmd` signature preserved)
- [x] 4.8 Verify build, tests, lint

## 5. OpenSpec change + doc sync (Stage 5)
- [x] 5.1 `openspec new change sandbox-escape-hardening-fixes`
- [x] 5.2 Write `proposal.md` with Why / What Changes / Capabilities / Impact
- [x] 5.3 Write `design.md` with Context / Goals / Non-Goals / Decisions D1-D4 / Risks / Migration Plan
- [x] 5.4 Write delta spec `specs/os-sandbox-core/spec.md` (MODIFIED: Policy types with isDir guard scenarios; Seatbelt profile generation with read+write deny scenarios)
- [x] 5.5 Write `tasks.md` (this file)
- [x] 5.6 `openspec validate sandbox-escape-hardening-fixes --strict`
- [x] 5.7 In-skill verification — 0 CRITICAL findings
- [x] 5.8 Sync delta into `openspec/specs/os-sandbox-core/spec.md`
- [x] 5.9 `openspec archive` → move change to `openspec/changes/archive/YYYY-MM-DD-sandbox-escape-hardening-fixes/`
- [x] 5.10 Verify README / docs / prompts do not need changes (PR 4 docs already describe the correct contract; this PR makes the code match)
- [x] 5.11 Final verify: build cross-platform, test, lint
