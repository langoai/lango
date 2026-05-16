## Why

Once a dedicated agent CLI reference existed, `docs/cli/core.md` still embedded a second copy of the agent diagnostics command manual. That duplicated public CLI docs and made future agent-trace updates easier to miss in one of the two places.

## What Changes

- add `docs/cli/agent.md` as the dedicated agent diagnostics CLI reference
- remove the duplicated embedded agent diagnostics manual from `docs/cli/core.md`
- add an executable guard so core docs keep delegating those commands to the dedicated agent page

## Impact

- less duplication across public CLI references
- agent diagnostics now have a single source of truth
- stronger regression protection for CLI docs scope drift
