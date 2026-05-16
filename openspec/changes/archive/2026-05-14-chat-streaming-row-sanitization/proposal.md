## Why

Finalized assistant content is sanitized before markdown rendering, but the in-flight streaming row still renders `streamBuf` raw. Malformed streamed chunks can therefore leak terminal control sequences into the transcript during generation.

## What Changes

- Sanitize in-flight streaming transcript content before rendering it.
- Add regression coverage for escaped streamed chunks.
- Record the streaming-row sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Brings the live streaming surface up to the same plain-text baseline as finalized assistant content.
- Prevents malformed streamed output from destabilizing the transcript mid-turn.
