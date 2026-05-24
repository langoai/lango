## Why

The chat transcript already claims approval events carry compact request-id annotations, but the implementation still appends the full request ID string. Real runtime request IDs are long enough to contradict that contract directly.

## What Changes

- Compact long approval request IDs before rendering them into approval transcript events.
- Preserve short IDs as-is.
- Add regressions and record the compaction contract in OpenSpec.

## Impact

- Aligns the approval transcript with the documented compact-ID contract.
- Improves readability for approval-heavy chat sessions.
