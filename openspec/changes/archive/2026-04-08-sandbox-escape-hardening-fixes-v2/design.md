## Context

`sandbox-escape-hardening-fixes` landed on 2026-04-08 and closed four concrete regressions that the first Codex review flagged after PR 4 archived. A second Codex review pass immediately after surfaced two more issues that the first fix PR did not address, plus two design-level gaps that belong in PR 5. This change (`sandbox-escape-hardening-fixes-v2`) picks up the two remaining small fixes and defers the two design-level items.

The bwrap mount order bug is older than PR 4 — it has shipped on Linux bwrap since the backend first landed in PR 3. It was never caught because the existing unit tests only verified that bwrap wraps the command correctly and that the basic argv contains the expected flags, not that the flags appear in the order bubblewrap actually needs. The second Codex pass specifically read bubblewrap's option-processing semantics and noticed the ordering problem.

The sandbox path overlap bug was introduced by Round 1's expansion of `DefaultToolPolicy` to deny `cfg.DataRoot`. Users who set a relative sandbox path get it normalized under `cfg.DataRoot`, then the same dataRoot gets added to DenyPaths, and the deny tmpfs covers the workspace. The failure is silent: the user sees write permission errors at runtime with no configuration-level diagnostic.

## Goals / Non-Goals

**Goals:**
- bwrap sandboxed children see a fresh `/proc` (new PID namespace), filtered `/dev`, and empty `/run`, not the host's versions.
- Users with configurations that would silently break the workspace see an actionable error at startup instead.
- No new features. No schema migrations. No TUI changes. No changes outside `internal/sandbox/os/bwrap_args.go` and `internal/config/loader.go`.

**Non-Goals (deferred to PR 5):**
- **Subdirectory walk-up `.git` discovery.** `DefaultToolPolicy` currently only checks `filepath.Join(workDir, ".git")`. When lango is started from `repo/subdir`, the parent repository's `.git` is not protected. A proper fix requires a git-aware policy builder that walks up from `workDir` to find the repository root. This expands the escape surface (a malicious workspace could rewrite `.git` as a symlink to a sensitive target, tricking the sandbox into denying unrelated paths) and is grouped with the PR 5 backend robustness work.
- **`MCPServerPolicy` workspace `.git` deny.** `MCPServerPolicy` only denies `dataRoot`; the workspace `.git` is never masked for stdio MCP server children. Adding it requires passing a workspace path through `ServerConnection.SetOSIsolator` and setting `cmd.Dir` in `createTransport`. This is PR 5 scope.
- `bwrap --version` probe false positives on hosts with `kernel.unprivileged_userns_clone=0`.
- Native Landlock+seccomp backend; file-level deny; symlink chain resolution; glob/path semantics normalization; per-tool policy overrides.

## Decisions

### D1 — bwrap mount order: root bind first, specialised mounts second

**Problem**: `compileBwrapArgs` emitted `--proc /proc`, `--dev /dev`, and `--tmpfs /run` as part of the initial base args slice, then later appended `--ro-bind / /`. bubblewrap processes options left-to-right: a later `--ro-bind / /` creates a new mount over `/` that shadows anything mounted under the sandbox root earlier. Every bwrap invocation therefore ran with the host's `/proc` visible inside the sandbox, defeating the `--unshare-pid` PID namespace (child could see host PIDs via `/proc/N`) and weakening `/dev` isolation.

**Decision**: Reorder `compileBwrapArgs` so the specialised mounts are appended AFTER the root/read bind. The new order is:
1. Namespace flags (`--die-with-parent`, `--unshare-pid`, ...)
2. Root bind (`--ro-bind / /`) or explicit ReadPaths
3. `--proc /proc`, `--dev /dev`, `--tmpfs /run`
4. Write binds
5. Deny tmpfs mounts
6. Network unshare

This matches the canonical pattern used by flatpak and every other bwrap wrapper. The specialised mounts now sit on top of the root bind, so the sandboxed child sees a fresh procfs, a filtered devtmpfs, and an empty /run.

**Regression guard**: `TestCompileBwrapArgs_RootBindBeforeSpecialMounts` asserts the argv index order directly (`--ro-bind / /` index must be less than `--proc /proc`, `--dev /dev`, `--tmpfs /run` indices). This is a cheap, deterministic check that catches any future reshuffling of the base args slice.

**Why not test runtime enforcement**: Would require running bwrap in CI and inspecting `/proc/1/status` inside the sandbox. That depends on kernel capabilities (unprivileged userns) that may not be present in every runner. The argv order check is a sufficient proxy — if bwrap receives the right arguments in the right order, the kernel handles the rest.

### D2 — `config.Validate` rejects sandbox paths under `cfg.DataRoot`

**Problem**: `NormalizePaths` resolves relative `sandbox.workspacePath` and `sandbox.allowedWritePaths` entries against `cfg.DataRoot`. A user writing `sandbox.workspacePath: repo` ends up with `/home/user/.lango/repo`. `DefaultToolPolicy` then adds `cfg.DataRoot` to `DenyPaths`, which in bwrap is a `--tmpfs ~/.lango` mount that comes AFTER the `--bind ~/.lango/repo ~/.lango/repo` write mount. The later deny tmpfs covers the entire workspace, making it unreachable from the sandboxed child. Seatbelt has the same failure mode — deny rules are rendered after allow rules in the profile template and Seatbelt evaluates last-match-wins. The user sees writes silently fail without any clear error pointing at the config.

**Decision**: Add a post-normalization check in `config.Validate` that rejects any `sandbox.workspacePath` or `sandbox.allowedWritePaths` entry that resolves to the same path as `cfg.DataRoot` or to a subtree of it. The error message is actionable: it names the colliding path, the DataRoot, and tells the user to use an absolute path outside the control-plane.

**Why validate-and-reject instead of rewriting the normalization base to cwd**:
1. Keeps `NormalizePaths` behavior consistent for all path fields (everything resolves against DataRoot by default).
2. Catches the problem for both relative-that-got-normalized-under-DataRoot AND for absolute-but-user-typed-an-inside-path cases. Both are the same runtime bug, both get the same clear error.
3. Changing the relative base to `os.Getwd()` for sandbox fields only would diverge sandbox semantics from every other data path, creating a different kind of confusion for users who understand the current convention.
4. Explicit rejection at startup is better than silent runtime breakage. Users get a message, users fix the config.

**Helper**: `pathIsUnder(child, parent string) bool` uses `filepath.Rel` and treats `rel == "."` as "same path" (true), `rel == ".."` or `.."/..."` as "outside" (false), and everything else as "nested" (true). Returns false on empty inputs and on `filepath.Rel` errors (different volumes on Windows). Unit-tested with eight cases covering nested, same, sibling, parent-is-child, trailing separator, and empty inputs.

## Risks / Trade-offs

- **Reordering bwrap args is a runtime-visible change on Linux**. Mitigation: the existing unit tests for bwrap write, read, workspace write, and network deny smoke cases still pass with the new order (verified). The argv order test is the new regression guard.
- **Relative sandbox path users get a hard startup error** where they previously got a silent runtime failure. This is strictly an improvement (explicit > silent), but users with working setups that happen to collide will need to update their config to use an absolute path outside `cfg.DataRoot`.
- **Tests do not verify runtime bwrap enforcement** — they assert argv order only. This is intentional (see D1). Runtime verification is still covered by `lango sandbox test`.

## Migration Plan

No schema migration, no config migration. On first run after deployment:
- Linux bwrap invocations pick up the new argv order automatically.
- Users with sandbox paths inside `cfg.DataRoot` see a startup error with a clear fix (move the path outside DataRoot). This may require manual config changes for affected users.

Rollback: revert the two commits. No data to undo.
