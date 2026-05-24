## Why

Mission Control is already durable-first, but its default first-screen composer hint still says `Type to chat here`. That wording understates the actual surface: the composer is for top-level requests that normally start durable mission work, not just generic chat.

## What Changes

- Change the default Mission Control composer hint from chat-only wording to neutral request wording.
- Update the workbench footer/helper copy that reused the same stale phrase.
- Sync cockpit feature docs and downstream docs requirements.

## Impact

- First-screen Mission Control copy better matches the durable-first runtime contract.
- Workbench and cockpit hints stop nudging operators toward a chat-only mental model.
