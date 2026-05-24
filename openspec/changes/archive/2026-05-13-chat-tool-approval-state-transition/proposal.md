## Why

The chat transcript already defines `awaiting_approval` and `canceled` tool item states, but the approval flow does not actually drive tool rows through those states. A tool row stays `running` while approval is pending, and a denied request never marks the row canceled.

## What Changes

- Move the latest matching running tool row into `awaiting approval` when an approval request arrives.
- Restore that row to `running` on approval and mark it `canceled` on denial.
- Keep the compact param preview visible through those transitions.
- Add regressions and record the contract in OpenSpec.

## Impact

- Makes the tool lifecycle transcript reflect the real approval state machine.
- Improves operator traceability during approval interruptions and denials.
