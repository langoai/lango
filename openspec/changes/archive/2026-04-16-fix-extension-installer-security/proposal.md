# Proposal: Fix Extension Installer Security & Accuracy

## Problem

Codex review (round 2, base: main) found 3 issues in the extension installer subsystem:

1. **P1 TOCTTOU symlink escape**: `copyTree()` walks skill directories without per-file `ResolvePath` validation. Between Inspect and Install, symlinks can be replaced to escape the pack root.
2. **P2 commit-pinned sources fail**: `git clone --branch <ref>` only works for branch/tag names, not commit SHAs. The documented `repo.git#abc1234` format fails at clone time.
3. **P2 inspect preview underreports**: `plannedWrites()` only lists the manifest `SKILL.md` path. Install copies the full skill directory tree, so sibling resources are not shown in the inspect preview.

## Proposed Solution

- Fix 1: Add `rootDir` parameter to `copyTree`, call `ResolvePath` per file in the Walk callback. Add `os.Lstat` symlink rejection in `copyFile` as belt-and-suspenders.
- Fix 2: Detect SHA-like refs, use clone-without-branch + `git checkout` strategy for commit pinning.
- Fix 3: Expand `plannedWrites` to walk skill directories and enumerate all files.

## Impact

- Fix 1 closes a P1 security vulnerability (symlink escape during install)
- Fix 2 enables commit-pinned extension sources as documented in spec
- Fix 3 restores the inspect-before-confirm trust model for multi-file skill packs
