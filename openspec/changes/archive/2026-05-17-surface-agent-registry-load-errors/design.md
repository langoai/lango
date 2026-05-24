# Design

## Root Cause

`agentregistry.FileStore.Load` and `Registry.LoadFromStore` already return read and parse errors. The CLI inspection commands drop those errors at their call sites, which turns a real registry problem into incomplete output.

## Approach

- Change `agent list` to return `load user agents: ...` when `agent.agentsDir` is configured and the store returns an error.
- Change the status registry counter helper to return `(counts, error)` so `agent status` can fail visibly on embedded or user store errors.
- Keep missing user agent directories optional because `FileStore.Load` already treats `os.ErrNotExist` as no definitions.
- Add focused tests for invalid configured `AGENT.md` files in both list and status commands.

## Downstream Impact

This is a CLI behavior change only. No runtime agent routing behavior changes are intended. Docs should explain that missing `agent.agentsDir` is optional, but invalid present definitions fail visibly.
