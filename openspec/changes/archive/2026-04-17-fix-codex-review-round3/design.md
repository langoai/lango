# Design: Codex Review Round 3 Fixes

## Fix 1: Wire runtime KnowledgeSaver

- Add `SaveToolResult(ctx, sessionKey, toolName, params, result) error` to `knowledge.Store`
- Method wraps the call into `SaveKnowledge` with appropriate `KnowledgeEntry` construction
- In `app.go` line ~177: extract `iv.KC.store` when `iv != nil && iv.KC != nil` and assign to `knowledgeSaver`
- Compile-time interface check: `var _ toolchain.KnowledgeSaver = (*knowledge.Store)(nil)`

## Fix 2: Resolve symlinks before copyFile in copyTree

- After `ResolvePath` validation passes in `copyTree` Walk callback, call `filepath.EvalSymlinks(path)` to resolve the symlink
- Pass the resolved path to `copyFile` instead of the raw Walk path
- This matches Inspect behavior: ResolvePath validates containment, then the resolved path is used for the actual operation
- Update `TestCopyFileRejectsSymlink` to only test external symlinks (not in-root ones)
- Add `TestCopyTreeAcceptsInRootSymlink` test

## Fix 3: Propagate Version to all bootstrap.Run calls

- `doctor.go`, `settings.go`, `onboard.go` all import `cliboot` and pass `Version: cliboot.Version`
- Alternative: have bootstrap.Options default-fill from cliboot.Version. But explicit is better.
