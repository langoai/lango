## Why

The `lango graph` command family was implemented and exposed in public quick references, but unlike the other major CLI families it still lacked a dedicated reference page under `docs/cli/`. That made the graph command surface harder to inspect and easier to document inconsistently.

## What Changes

- add `docs/cli/graph.md` covering the implemented `status`, `query`, `stats`, `clear`, `add`, `export`, and `import` commands
- add an executable guard so the graph CLI reference keeps documenting the current command and flag contract

## Impact

- dedicated graph docs now match the actual shipped CLI surface
- graph output, import/export, and destructive-clear behavior are easier for operators and wrappers to understand
- stronger regression protection for graph docs drift
