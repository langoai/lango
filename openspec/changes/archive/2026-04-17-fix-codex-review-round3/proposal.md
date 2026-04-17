# Proposal: Codex Review Round 3 Fixes

## Problem

Three findings from Codex review (base: dev):

1. **P1**: `buildHookRegistry` always passes `nil` for `knowledgeSaver`. The hook is registered but never persists tool results. `knowledge.Store` needs to implement `toolchain.KnowledgeSaver`.
2. **P2**: `copyFile` rejects ALL symlinks via `os.Lstat`, but `fetchFromDir` allows in-root symlinks via `ResolvePath`. Creates inspect-succeeds/install-fails inconsistency.
3. **P2**: `lango doctor`, `lango settings`, `lango onboard` call `bootstrap.Run(bootstrap.Options{})` without Version, producing empty version in timing logs.

## Proposed Solution

1. Add `SaveToolResult` method to `knowledge.Store` as adapter wrapping `SaveKnowledge`. Wire `iv.KC.store` through `buildHookRegistry`.
2. In `copyTree`, use resolved path from `filepath.EvalSymlinks` when calling `copyFile`, matching the Inspect path behavior.
3. Pass `cliboot.Version` to all `bootstrap.Run` call sites.
