## Why

Chat approval and tool-preview surfaces already sanitize parameter values, but parameter keys are still rendered raw. A malformed external schema or tool payload can still inject control sequences or embedded newlines through the key names themselves.

## What Changes

- Sanitize displayed parameter keys across approval banner, approval dialog, and tool param preview surfaces.
- Add regression coverage for escaped and multiline parameter keys.
- Record the parameter-key sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Completes the plain-text rendering baseline for parameter display, not just parameter values.
- Prevents malformed external parameter schemas from destabilizing approval/tool preview UI.
