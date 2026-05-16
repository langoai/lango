## Why

The cockpit Dead Letters page renders backlog reasons, transaction identifiers, actor/dispatch summaries, detail metadata, and retry follow-up messages directly from runtime-fed post-adjudication data. Those values currently pass through without normalization, so malformed values can leak control sequences or embedded newlines into a read-only operator surface.

## What Changes

- Sanitize operator-facing Dead Letters text before rendering it.
- Sanitize retry success/failure/follow-up status messages.
- Add regressions for malformed dead-letter backlog, detail, and status metadata.
- Record the plain-text Dead Letters contract in OpenSpec and downstream docs.

## Impact

- Brings the cockpit Dead Letters page up to the same plain-text baseline as the hardened chat and other cockpit surfaces.
- Prevents malformed post-adjudication metadata from destabilizing the operator dashboard.
