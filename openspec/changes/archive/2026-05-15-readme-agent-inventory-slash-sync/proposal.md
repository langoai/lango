## Why

The README internal CLI inventory still describes the agent diagnostics slice with stale hyphen-compressed shorthand: `trace list-show-metrics/graph`. The current command surface and architecture inventory already use the clearer slash-separated form `trace list/show/metrics/graph`.

## What Changes

- update the README internal tree agent row to the current slash-separated diagnostics wording
- update the existing A2A/agent inventory guard to reject the stale hyphen shorthand
- sync the main docs-only and test-coverage specs with that README contract

## Impact

- more truthful public inventory wording
- less ambiguity about the actual agent diagnostics command paths
- stronger regression protection against stale README shorthand
