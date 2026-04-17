# Design: Fix Extension Installer Security & Accuracy

## Fix 1: TOCTTOU symlink escape in copyTree

### Current State
- `fetchFromDir()` in source.go re-validates each file via `hashFile()→ResolvePath()` during Inspect — **safe**
- `copyTree()` in installer.go walks via `filepath.Walk` and calls `copyFile(path, target)` without re-validation — **vulnerable**
- `copyFile()` uses `os.Open()` which follows symlinks — no rejection

### Approach
- `copyTree(src, dst, rootDir string) error` — add rootDir parameter
- Walk callback: compute `filepath.Rel(rootDir, path)`, call `ResolvePath(rootDir, rel)` per file
- `copyFile`: add `os.Lstat(src)` check, reject `ModeSymlink`
- Update both callers: `copyPackFiles()` line 266 and `copySkillsToStore()` line 302

### Decision: ResolvePath + Lstat double-check
ResolvePath validates containment via canonical path resolution. Lstat rejects raw symlinks before Open. Both are needed because ResolvePath alone has a race window between resolve and open.

## Fix 2: Commit-pinned extension sources

### Current State
- `git clone --branch <ref>` at source.go:86-88
- Only works for branch/tag names; commit SHAs fail with "Remote branch abc1234 not found"

### Approach
- Add `looksLikeSHA(ref) bool` — 7-40 hex character check
- SHA path: clone without `--branch` and without `--depth=1` (shallow clone can't fetch arbitrary SHAs), then `git -C tmp checkout <ref>`
- Branch/tag path: existing `--branch` + `--depth=1` behavior unchanged
- Extract clone+checkout logic into `resolveGitRef()` helper for testability

## Fix 3: Inspect preview underreports

### Current State
- `plannedWrites(m *Manifest)` iterates `m.Contents.Skills` and reports `s.Path` only
- `Install()` converts SKILL.md paths to parent directories and copies full directory trees

### Approach
- `plannedWrites(m *Manifest, rootDir string)` — add rootDir parameter
- For each skill: convert SKILL.md to parent dir (same as copyPackFiles), stat, if directory then Walk and enumerate all files
- Update Inspect call site to pass `wc.RootDir`
