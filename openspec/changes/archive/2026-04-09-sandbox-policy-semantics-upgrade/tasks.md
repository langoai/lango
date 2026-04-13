## 1. File-level deny (D1)

- [x] 1.1 Rewrite `compileBwrapArgs` DenyPaths loop as `switch mode`: directory → `--tmpfs`, regular file → `--ro-bind /dev/null`, other → error
- [x] 1.2 Update doc comment in `bwrap_args.go` (PR 4 planned → PR 5c implemented)
- [x] 1.3 Delete `TestCompileBwrapArgs_DenyPathMustBeDirectory`
- [x] 1.4 Add `TestCompileBwrapArgs_DenyPathFileGetsRoBindDevNull`
- [x] 1.5 Add `TestCompileBwrapArgs_DenyPathDirectoryStillGetsTmpfs` (regression guard)
- [x] 1.6 Add `TestCompileBwrapArgs_DenyPathUnsupportedMode` using portable `/dev/null` character device

## 2. Symlink resolution + glob expansion via shared `normalizePath` (D2 + D3)

- [x] 2.1 Add `normalizePath(entry string) ([]string, error)` helper in `policy.go` — canonical pipeline (sanitize → Abs → Glob → EvalSymlinks-with-fallback)
- [x] 2.2 Add `canonicalWorkDir(workDir)` helper (Abs + EvalSymlinks + fallback)
- [x] 2.3 Add `collectBaselineDeny(workDir, dataRoot)` helper implementing the two-deny strategy
- [x] 2.4 Rewrite `findGitRoot` to return `gitRoot` struct with `pointerPath` + `gitdirPath`
- [x] 2.5 Add `parseWorktreePointer` helper — reads first line, parses `gitdir:`, resolves relative or absolute target
- [x] 2.6 Refactor `DefaultToolPolicy`/`MCPServerPolicy` to use `canonicalWorkDir` + `collectBaselineDeny`
- [x] 2.7 Update `compileBwrapArgs` — ReadPaths/WritePaths/DenyPaths loops call `normalizePath` (returning slices)
- [x] 2.8 Update `seatbelt_profile.go` — same three loops use `normalizePath`

## 3. Test coverage

- [x] 3.1 Add `resolveSymlinks(t, p)` helper in `policy_test.go` for macOS canonical-path expectations
- [x] 3.2 Update existing `TestDefaultToolPolicy*`, `TestMCPServerPolicy*`, `TestStrictToolPolicy`, `TestGenerateSeatbeltProfile` with resolveSymlinks
- [x] 3.3 Rewrite `TestFindGitRoot` for new `gitRoot` struct — add worktree absolute/relative gitdir, malformed pointer, symlinked workDir cases
- [x] 3.4 Add `TestDefaultToolPolicy_WorktreeDenyBothPointerAndGitdir`
- [x] 3.5 Rewrite `TestDefaultToolPolicy_GitFileNotDenied` → malformed-pointer pointer-only deny
- [x] 3.6 Add `TestCompileBwrapArgs_SymlinkedDenyPath` (symlink escape closed)
- [x] 3.7 Add `TestNormalizePath_NonexistentFallback`
- [x] 3.8 Add `TestNormalizePath_GlobExpansion`
- [x] 3.9 Add `TestNormalizePath_UnmatchedGlobSilentSkip`
- [x] 3.10 Add `TestNormalizePath_InvalidGlobErrors`
- [x] 3.11 Add `TestCompileBwrapArgs_DenyPathWithGlob` (integration)
- [x] 3.12 Update `TestCompileBwrapArgs_DefaultToolPolicy_GitFile` for new worktree semantics

## 4. Documentation sync

- [x] 4.1 `docs/configuration.md` — add "Path semantics (file-level deny, symlinks, globs)" paragraph; annotate `allowedWritePaths` row with shared pipeline note
- [x] 4.2 `docs/cli/sandbox.md` — add "Path semantics" paragraph to the experimental note
- [x] 4.3 `README.md` — extend the OS-level Sandbox bullet with file-deny, symlink, glob
- [x] 4.4 `prompts/SAFETY.md` — extend the Control-plane bullet with symlink/worktree/glob mentions

## 5. OpenSpec change

- [x] 5.1 `openspec new change sandbox-policy-semantics-upgrade`
- [x] 5.2 Write `proposal.md`
- [x] 5.3 Write `design.md`
- [x] 5.4 Write this `tasks.md`
- [x] 5.5 Write delta spec for `linux-bwrap-isolation` (MODIFY DenyPaths requirement to cover file support)
- [x] 5.6 Write delta spec for `os-sandbox-core` (MODIFY Policy types requirement to cover normalizePath + gitRoot + two-deny strategy)
- [ ] 5.7 Run `openspec validate sandbox-policy-semantics-upgrade --strict`
- [ ] 5.8 Run `openspec archive sandbox-policy-semantics-upgrade -y` (no `--no-validate` — PR 5b's success criterion)
- [ ] 5.9 Stop at commit boundary — user commits manually
