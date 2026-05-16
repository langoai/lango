## Why

Approval transcript events already compact request IDs and summaries, but they still only trim whitespace from the event text and summary before storing them. Malformed tool names or request summaries can still leak terminal control sequences into the transcript traceability surface.

## What Changes

- Sanitize approval transcript event text and request-summary preview before storing them.
- Add regression coverage for escaped and multiline approval-event text.
- Record the approval-event sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Keeps approval transcript traceability aligned with the rest of the sanitized approval surfaces.
- Prevents malformed approval metadata from destabilizing the transcript lane.
