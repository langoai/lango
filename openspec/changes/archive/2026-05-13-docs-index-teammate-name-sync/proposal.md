## Why

The public home page still describes the multi-agent system with legacy built-in teammate names: `Executor, Researcher, Planner, Memory Manager`.

The runtime roster has long since moved to the current built-in teammate set (`operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, `ontologist`). The landing page should not advertise the wrong names.

## What Changes

- Update `docs/index.md` to use current built-in teammate examples.
- Extend downstream docs requirements so the home page is covered by the same current-name contract as the other public docs.

## Impact

- The public landing page reflects the actual built-in teammate system.
- Downstream docs sync no longer leaves `docs/index.md` outside the current-name truth contract.
