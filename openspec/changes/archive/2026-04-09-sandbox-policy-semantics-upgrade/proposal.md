## Why

PR 5a closed the immediate contract gaps in bwrap probe robustness, `.git` walk-up, and MCP workspace asymmetry, but three path-semantic gaps remained:

1. **File-level deny was explicitly unsupported**: `compileBwrapArgs` rejected every regular file in `DenyPaths` with "must be a directory; file-level deny not yet supported". This blocked denying individual secret files like `~/.lango/lango.db`, encrypted config tokens, and linked-worktree `.git` pointer files.
2. **No symlink resolution**: `EvalSymlinks` was used nowhere in the sandbox package. A symlinked `.git` directory (or any symlinked deny path) would cause `--tmpfs` to mount on the symlink itself, leaving the real target fully readable — a silent symlink escape.
3. **No glob expansion**: `Sandbox.AllowedWritePaths` entries containing `*` were treated as literal strings. Users writing `~/.lango/*.db` saw no match, no error, and no effect.

These gaps all live in the same path-handling layer and all three backends (bwrap, Seatbelt, planned native Linux) need identical semantics to avoid drift. PR 5c introduces a canonical path normalization pipeline that every backend shares, closing all three gaps in one causally-linked change.

## What Changes

- **NEW**: `normalizePath(entry)` helper in `internal/sandbox/os/policy.go` — the shared 6-step pipeline `sanitize → Abs → Glob → EvalSymlinks (with nonexistent fallback) → []string` used by every policy path consumer.
- **BEHAVIOR CHANGE**: `compileBwrapArgs` now accepts regular files in `DenyPaths`. Directories still emit `--tmpfs <path>`; regular files emit `--ro-bind /dev/null <path>`. Device nodes, sockets, and fifos surface as errors.
- **BEHAVIOR CHANGE**: `compileBwrapArgs` ReadPaths/WritePaths/DenyPaths loops call `normalizePath` instead of `sanitizePath` directly, so globs expand and symlinks resolve for all path classes.
- **BEHAVIOR CHANGE**: `GenerateSeatbeltProfile` routes all path classes through the same `normalizePath` — bwrap and Seatbelt see entries in identical shape.
- **BEHAVIOR CHANGE**: `findGitRoot` now returns a `gitRoot` struct with `pointerPath` + `gitdirPath`. For standard repos, both equal the `.git` directory. For linked worktrees, `pointerPath` is the `.git` file (denied at file level via the PR 5c file-level deny) and `gitdirPath` is the parsed+resolved gitdir target. Malformed pointers degrade to pointer-only deny.
- **BEHAVIOR CHANGE**: `DefaultToolPolicy`/`MCPServerPolicy` use a new `canonicalWorkDir` helper (Abs + EvalSymlinks + fallback) so `WritePaths[0]` is the canonical filesystem path — symlinked workspaces no longer leak their pre-resolve path into the writable set. They share a `collectBaselineDeny` helper that applies the two-deny strategy.
- **BREAKING (internal)**: `findGitRoot` return type `string` → `gitRoot`. Call sites inside `policy.go` updated.
- **Tests**: new tests covering symlinked DenyPath, symlinked workDir walk-up, worktree pointer with absolute/relative gitdir, malformed pointer degradation, glob expansion (match/no-match/invalid pattern), `normalizePath` nonexistent fallback, file-level deny + directory deny cohabitation, and cross-backend symmetry.
- **Docs**: `docs/cli/sandbox.md`, `docs/configuration.md`, `README.md`, and `prompts/SAFETY.md` each gain a paragraph or line noting the new file-level deny, symlink resolution, and glob support.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `os-sandbox-core`: `Policy` types requirement updated — `findGitRoot` returns `gitRoot` struct, `DefaultToolPolicy`/`MCPServerPolicy` add two-deny strategy entries, shared `normalizePath` pipeline contract documented.
- `linux-bwrap-isolation`: `compileBwrapArgs` DenyPaths requirement updated — directories still use `--tmpfs`, regular files use `--ro-bind /dev/null`, other file modes error. ReadPaths/WritePaths/DenyPaths all flow through `normalizePath` for glob expansion and symlink resolution.

## Impact

**Code**:
- `internal/sandbox/os/policy.go` (normalizePath, gitRoot, parseWorktreePointer, canonicalWorkDir, collectBaselineDeny, DefaultToolPolicy/MCPServerPolicy refactor)
- `internal/sandbox/os/bwrap_args.go` (DenyPaths file/dir switch + normalizePath in all 3 loops + updated doc comment)
- `internal/sandbox/os/bwrap_args_test.go` (+7 new tests, updates to existing tests for resolved paths)
- `internal/sandbox/os/policy_test.go` (+5 new tests, updates to existing tests)
- `internal/sandbox/os/seatbelt_profile.go` (normalizePath in all 3 loops)
- `docs/cli/sandbox.md`, `docs/configuration.md`, `README.md`, `prompts/SAFETY.md` (user-facing doc updates)

**Behavior**:
- Users who had symlinked `.git` directories in their workspace are now correctly protected.
- Users who included glob patterns in `allowedWritePaths` see them expanded to concrete matches.
- Users who included individual secret files in custom deny lists see them actually denied (previously produced an error that rejected the whole policy).
- Linked git worktrees get protection for both the `.git` pointer file and the resolved gitdir target.

**Dependencies**: None new. Uses only `filepath.Glob` and `filepath.EvalSymlinks` from the standard library.
