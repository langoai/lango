## Why

The `lango account` surface advertises `--output table|json` across multiple subcommands, but those commands do not validate the flag before running bootstrap and smart-account setup work.

## What Changes

- validate `--output table|json` across smart-account subcommands before load or bootstrap work
- silence duplicate Cobra error and usage noise for these fail-fast validation errors
- add regressions covering representative root and nested subcommands
- sync smart-account docs and specs

## Impact

- more predictable automation and operator UX
- less wasted bootstrap work on invalid invocations
- tighter alignment between docs, tests, and implementation
