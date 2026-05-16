## Why

The `lango extension` command family was implemented and wired into the root CLI, but unlike the other major CLI families it still lacked a dedicated reference page under `docs/cli/`. That made the command surface harder to inspect and easier to document inconsistently.

## What Changes

- add `docs/cli/extension.md` covering the implemented `inspect`, `install`, `list`, and `remove` commands
- add an executable guard so the extension CLI reference keeps documenting the current output and confirmation contract

## Impact

- dedicated extension docs now match the actual shipped CLI surface
- extension output and confirmation behavior is easier for operators and wrappers to understand
- stronger regression protection for extension docs drift
