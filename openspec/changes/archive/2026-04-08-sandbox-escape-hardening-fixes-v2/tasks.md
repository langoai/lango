## 1. bwrap mount order (Stage 1)
- [x] 1.1 Reorder `compileBwrapArgs` in `internal/sandbox/os/bwrap_args.go` so `--proc /proc`, `--dev /dev`, `--tmpfs /run` are appended AFTER the `--ro-bind / /` (or explicit ReadPaths) block
- [x] 1.2 Update the function doc comment explaining why mount order is load-bearing (bubblewrap left-to-right processing, later root bind shadows earlier specialised mounts)
- [x] 1.3 Add `TestCompileBwrapArgs_RootBindBeforeSpecialMounts` asserting the argv index order (`--ro-bind / /` index < `--proc`, `--dev`, `--tmpfs /run` indices)
- [x] 1.4 Verify existing bwrap tests still pass (`TestCompileBwrapArgs_*`)
- [x] 1.5 Verify build cross-platform (`go build ./... && GOOS=linux GOARCH=amd64 go build ./...`), tests (`go test ./...`), lint (`golangci-lint run ./internal/sandbox/os/...`)

## 2. Sandbox path DataRoot overlap validation (Stage 2)
- [x] 2.1 Add `pathIsUnder(child, parent string) bool` helper in `internal/config/loader.go` (uses `filepath.Rel`, treats `.` as nested, `..[/...]` as outside, returns false on empty inputs)
- [x] 2.2 Add post-normalization check in `Validate` that rejects `sandbox.workspacePath` and every entry of `sandbox.allowedWritePaths` nested under `cfg.DataRoot` with an actionable error message naming the colliding path
- [x] 2.3 Extend `TestValidate` in `internal/config/loader_test.go` with 5 new subtests:
  - workspacePath under DataRoot rejected
  - workspacePath equal to DataRoot rejected
  - workspacePath outside DataRoot accepted
  - allowedWritePaths entry under DataRoot rejected
  - empty workspacePath accepted
- [x] 2.4 Add `TestPathIsUnder` with 8 cases (nested, same, sibling, parent-is-child, trailing separator, empty child, empty parent, outside)
- [x] 2.5 Verify existing `TestNormalizePaths_Sandbox` still passes
- [x] 2.6 Verify build cross-platform, tests, lint

## 3. OpenSpec change + sync + archive (Stage 3)
- [x] 3.1 `openspec new change sandbox-escape-hardening-fixes-v2` (renamed from -fixes to avoid collision with already-archived round-1 change)
- [x] 3.2 Write `proposal.md` with Why / What Changes / Capabilities / Impact (Round 2 only — Round 1 was already archived separately)
- [x] 3.3 Write `design.md` with Context / Goals / Non-Goals (P2#1 + P2#3 deferred to PR 5) / Decisions D1 (bwrap order) and D2 (sandbox path validation) / Risks / Migration Plan
- [x] 3.4 Write delta spec `specs/os-sandbox-core/spec.md` with `## ADDED Requirements` ONLY (Round 1's MODIFIED requirements are already in the main spec)
- [x] 3.5 Write `tasks.md` (this file)
- [x] 3.6 `openspec validate sandbox-escape-hardening-fixes-v2 --strict`
- [ ] 3.7 `openspec archive sandbox-escape-hardening-fixes-v2 -y` (native command — applies delta to main spec, moves change to archive directory)
- [ ] 3.8 Final verify: build cross-platform, test, lint
