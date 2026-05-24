## Why

Approval event transcript rows currently show only human text like `Approval requested for exec` or `Denied exec`. When the same tool is approved repeatedly in one session, the transcript does not clearly expose which approval request instance each event belongs to.

## What Changes

- Include a compact request-id annotation on approval transcript events when a request ID is available.
- Add regressions for the annotation.
- Record the traceability contract in OpenSpec and public docs.

## Impact

- Improves approval traceability in long-running chat sessions.
- Helps operators correlate transcript events with approval-history/request IDs.
