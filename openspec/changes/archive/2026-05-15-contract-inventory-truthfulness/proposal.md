## Why

The README internal CLI inventory still truncates the contract family to `abi`, even though the actual command path is `lango contract abi load`. That makes the public package tree less truthful than the dedicated CLI docs and the wired command surface.

## What Changes

- update the README internal tree to describe `lango contract read/call/abi load`
- add an executable inventory guard covering both `docs/architecture/project-structure.md` and the README internal tree
- sync the main docs-only and test-coverage specs with the current contract inventory contract

## Impact

- more truthful public inventory docs
- less risk of stale shorthand drifting back into the README tree
- stronger alignment between inventory docs and actual CLI wiring
