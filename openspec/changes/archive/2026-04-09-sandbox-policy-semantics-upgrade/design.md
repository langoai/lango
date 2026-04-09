## Design Notes

### D1: File-level deny via `--ro-bind /dev/null <file>`

**Problem**: `compileBwrapArgs` rejected every regular file in `DenyPaths` with `"bwrap deny path %q must be a directory; file-level deny not yet supported"`. This blocked the ability to deny individual secret files. Combined with PR 5a's inability to follow worktree `.git` pointers, it meant `~/.lango/lango.db` and linked-worktree `.git` files had no defense at the sandbox layer.

**Fix**: Replace the `!fi.IsDir()` rejection with a `switch` on `fi.Mode()`:
- `mode.IsDir()` → `--tmpfs <path>` (existing, empty tmpfs mask)
- `mode.IsRegular()` → `--ro-bind /dev/null <path>` (new — reads yield EOF, writes return EACCES)
- Other (device nodes, sockets, fifos) → error with `"unsupported file mode"` message

The `/dev/null` bind trick preserves the parent directory structure. The sandboxed child still sees the directory listing, still sees the filename, but opening the file gets redirected to `/dev/null` which is both read-empty and write-denied.

**Seatbelt (macOS) needs no code change**. The Seatbelt template's `(deny file-read* (subpath "<path>"))` + `(deny file-write* (subpath "<path>"))` rules already work on both directories and files because the `subpath` predicate operates on any filesystem path. Only bwrap needed the switch.

### D2: Symlink resolution via `filepath.EvalSymlinks`

**Problem**: Zero `EvalSymlinks` usage in the sandbox package. A symlinked `.git` directory or a symlinked deny path escaped protection because bwrap's `--tmpfs` mount lands on the symlink itself, not the resolved target.

**Fix**: Add `EvalSymlinks` to the canonical path pipeline at step 4 (after Abs, Glob, before Stat). For nonexistent paths (where EvalSymlinks errors with `os.IsNotExist`), fall back to the pre-resolve Abs path so downstream `os.Stat` catches the missing-path error with the existing error message format — the UX for "user typo'd a path" doesn't change.

`findGitRoot` also applies `EvalSymlinks` to `workDir` at entry so walk-up operates on the canonical filesystem path. A symlinked workDir (e.g. `/home/user/repo` → `/mnt/real/repo`) is resolved before the walk, and `.git` is discovered at the real ancestor path.

`DefaultToolPolicy`/`MCPServerPolicy` share a `canonicalWorkDir` helper that returns `filepath.Abs` + `filepath.EvalSymlinks` (with fallback). This makes `WritePaths[0]` the canonical filesystem path so bwrap `--bind` and Seatbelt `(allow file-write* (subpath ...))` both mount the real directory.

### D3: Glob expansion via `filepath.Glob`

**Problem**: Zero glob usage in the sandbox package. Users who wrote `~/.lango/*.db` in `allowedWritePaths` saw it treated as a literal string (no match, no error).

**Fix**: Step 3 of the canonical pipeline calls `filepath.Glob(absPath)` when `strings.ContainsAny(absPath, "*?[")`. Results:
- **Zero matches**: silent skip (shell `nullglob` semantics, aligns with the "best-effort baseline deny" philosophy throughout the sandbox code). `normalizePath` returns a nil slice with nil error.
- **One or more matches**: each match flows through the remaining pipeline (EvalSymlinks, Stat, type classification) independently. One entry in `DenyPaths` can expand to many `--ro-bind /dev/null` or `--tmpfs` emissions.
- **Invalid pattern** (`filepath.ErrBadPattern`, unclosed bracket, etc.): return a wrapped error. User config mistakes should surface at startup, not be silently dropped.

### Canonical Path Normalization Pipeline

```
entry → [1. sanitize] → [2. Abs] → [3. Glob] → [4. EvalSymlinks] → [5. Stat] → [6. type classify] → backend emission
```

Steps 1-4 live in the new `normalizePath` helper (policy.go). Steps 5-6 are backend-specific because bwrap distinguishes tmpfs/ro-bind while Seatbelt uses the same `(deny file-read* file-write* (subpath))` for both.

**Why Abs before Glob (settled)**: `filepath.Glob` accepts both relative and absolute patterns, but matching a relative pattern depends on cwd, which is unstable across call sites. Running `Abs` first turns every pattern into a cwd-rooted absolute string, then `Glob` matches against the filesystem deterministically.

**Why EvalSymlinks after Glob (settled)**: Glob patterns reference user-visible paths (which may contain symlinks). Each match is a concrete path that we then canonicalize via EvalSymlinks.

### Worktree `.git` file handling (two-deny strategy)

**Problem**: PR 5a's `findGitRoot` walked past linked worktree `.git` files because `compileBwrapArgs` rejected non-directory deny paths. With D1 lifting that restriction, we can follow the pointer.

**Chosen strategy**: deny BOTH the pointer file AND the resolved gitdir target.

`findGitRoot` now returns a `gitRoot` struct:
- `pointerPath`: the `.git` entry (file or directory) discovered via upward walk
- `gitdirPath`: the resolved gitdir target, which equals `pointerPath` for standard repos and the parsed `gitdir: <path>` target for linked worktrees (with relative paths resolved against the pointer file's parent)

`parseWorktreePointer` reads the first line of the `.git` file, parses `gitdir:`, and resolves the target via `Abs` + `EvalSymlinks`. On any failure (unreadable, missing prefix, empty target, invalid path) it degrades to `gitdirPath == ""` — callers still get file-level deny of the pointer file itself, which at least blocks direct reads of the gitdir pointer content.

`DefaultToolPolicy`/`MCPServerPolicy` add both paths to `DenyPaths` (when distinct). For standard repos that's one entry; for worktrees that's two (pointer file + resolved gitdir target, which may lie OUTSIDE the workspace — that's normal for worktrees created via `git worktree add`).

**Cross-filesystem concern**: The gitdir target for a worktree is typically at `/home/user/.git/worktrees/<name>`, outside the workspace. We deny it regardless — the backend just makes the path inaccessible to the child, and deny paths are allowed to reference anywhere on the filesystem.

### Cross-backend semantic alignment

The shared `normalizePath` helper guarantees bwrap and Seatbelt see entries in identical shape. The per-type translation table:

| Type | bwrap | Seatbelt | Native (future PR 5d) |
|------|-------|----------|-----------------------|
| Directory | `--tmpfs <path>` | `(deny file-read* (subpath "<path>"))` + `(deny file-write* (subpath "<path>"))` | Landlock: omit `path_beneath` rule → default-deny |
| Regular file | `--ro-bind /dev/null <path>` | Same as directory — Seatbelt `subpath` already handles files | Landlock ABI 3+: per-file path_beneath with no flags |
| Device/socket/fifo/other | Error — "unsupported file mode" | Error — same | Error — same |

**Cross-backend guarantees**:
- A deny on `~/.lango/lango.db` (file) → read AND write denied in both bwrap and Seatbelt
- A deny on `/repo/.git` (directory) → directory AND all descendants denied
- A symlinked path → denied at the resolved target, not the symlink
- A glob pattern → each match independently denied

## Alternatives Considered

- **Glob expansion at config load time, not sandbox construction**: would leak into `internal/config/loader.go`, coupling config to sandbox internals. Rejected — keep glob logic in the sandbox package where it belongs.
- **Follow worktree pointer recursively** (resolve submodule chain): out of scope for PR 5c. A `gitdir:` pointer target can itself contain a submodule structure, but the standard `git worktree` output doesn't nest them. Deferred to future work if a real need arises.
- **Parse worktree pointer via `git rev-parse --git-dir` subprocess**: requires subprocess execution inside Policy construction, violates the pure-function pattern, adds a git binary dependency. Rejected — stable `gitdir: <path>` format can be parsed in-process with 10 lines of Go.
- **Skip EvalSymlinks when all paths are absolute already**: false optimization. The user-visible path may contain symlinks at intermediate components (`/home/user/link → /mnt/real/home`), so EvalSymlinks on an already-absolute path is still meaningful.

## Worktree trade-off closing

PR 5a documented that linked worktrees retained the gap. PR 5c closes it via the two-deny strategy: pointer file denied at file level, gitdir target denied at directory level. Users with worktrees now get the same protection as users with standard repos.
