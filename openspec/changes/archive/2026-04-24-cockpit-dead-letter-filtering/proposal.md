## Why

The cockpit dead-letter page is already landed as a read-only master-detail surface, but operators still need a minimal way to narrow the backlog without dropping back to raw tool calls.

## What Changes

- add a thin cockpit filter bar with `query` and `adjudication`
- keep apply semantics simple: `Enter` reload + first-row reset
- document the cockpit filtering slice in public docs and main OpenSpec specs

## Impact

- better cockpit usability
- no new backend endpoints
- no live filtering or write actions in this slice
