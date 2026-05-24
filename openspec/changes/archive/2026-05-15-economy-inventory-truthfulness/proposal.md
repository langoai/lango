## Why

The README internal tree and architecture inventory still describe the economy family as broad buckets like `budget/risk/pricing/negotiate/escrow`, even though the actual CLI surface is status-oriented and the escrow slice includes `status`, `list`, `show`, and `sentinel status`.

## What Changes

- update the README internal tree economy row to reflect the current status-oriented command surface
- update the architecture `cli/economy/` row to the same current surface
- add an executable inventory guard covering both documents
- sync the main docs-only and test-coverage specs

## Impact

- more truthful public inventory docs
- better discoverability of the actual economy command paths
- stronger regression protection against stale shorthand returning
