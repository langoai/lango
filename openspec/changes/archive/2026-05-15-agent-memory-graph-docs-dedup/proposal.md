## Why

Once `docs/cli/graph.md` existed, `docs/cli/agent-memory.md` still embedded a full second copy of the graph command manual. That duplicated content across public CLI docs and made future graph updates easier to miss in one of the two places.

## What Changes

- remove the duplicated embedded graph command manual from `docs/cli/agent-memory.md`
- replace it with a direct handoff to `docs/cli/graph.md`
- add an executable guard so graph command duplication does not creep back into the agent-and-memory page

## Impact

- less duplication across public CLI references
- graph command updates now have a single source of truth
- stronger regression protection for CLI docs scope drift
