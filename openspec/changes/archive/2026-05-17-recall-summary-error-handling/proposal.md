## Why

Session recall can find a matching prior session but fail to load the stored summary for that row. The current adapter ignores that summary-load error and can return a recall match with an empty summary, which degrades the turn context while hiding the underlying recall backend failure.

## What Changes

- Add focused tests for recall adapter summary-load failures.
- Return an actionable error when `GetSummary` fails for a search hit.
- Preserve existing current-session filtering, rank-floor filtering, and top-N behavior.

## Impact

- Recall failures become observable through the existing `ContextAwareModelAdapter` warning path.
- LLM context no longer receives empty recall matches caused by hidden summary-load failures.
- No schema or configuration changes.
