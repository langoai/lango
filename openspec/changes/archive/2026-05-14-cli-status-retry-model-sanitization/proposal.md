## Why

The `lango status` table renderers and collected dashboard model are now sanitized, but dead-letter retry result models still store raw `message`, `follow_up_error`, subtype/family/reason/dispatch fields, and background-task status text. That leaves JSON output and any downstream reuse dependent on renderer-level sanitization instead of the retry result model itself being display-safe.

## What Changes

- Sanitize dead-letter retry result and follow-up model text at construction time.
- Add regression coverage for malformed retry-result JSON fields.
- Record the retry-result replay-safety contract in OpenSpec and downstream docs.

## Impact

- Extends status hardening from table rendering into the dead-letter retry JSON/result path.
- Prevents raw control text from persisting inside retry-result payloads returned to operators.
