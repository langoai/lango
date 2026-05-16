## Why

Once a dedicated config CLI reference existed, `docs/cli/core.md` still embedded a second copy of the config command manual. That duplicated public CLI docs and made future config updates easier to miss in one of the two places.

## What Changes

- remove the duplicated embedded config command manual from `docs/cli/core.md`
- replace it with a direct handoff to `docs/cli/config.md`
- add an executable guard so core docs keep delegating config coverage to the dedicated page

## Impact

- less duplication across public CLI references
- config command updates now have a single source of truth
- stronger regression protection for CLI docs scope drift
