## Why

Approval transcript events now carry compact request IDs, but the event text itself still only names the tool. When the same tool is approved for different targets or payloads, the transcript still makes operators jump elsewhere to recover the request summary.

## What Changes

- Append a compact request summary preview to approval transcript events when available.
- Keep the summary single-line-safe and concise.
- Add regressions and record the contract in OpenSpec/docs.

## Impact

- Improves approval-event traceability directly in the transcript.
- Reduces the need to cross-reference another surface just to understand which request was approved or denied.
