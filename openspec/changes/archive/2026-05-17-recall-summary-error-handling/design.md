## Overview

`recallProviderAdapter` adapts the recall backend to `adk.RecallProvider`. It already returns search errors to the caller, where the context adapter logs a warning and continues without recall. Summary retrieval errors should follow the same degraded-but-visible path instead of producing empty recall context.

## Decisions

### Fail the Recall Batch on Summary Load Error

If a candidate survives current-session and rank filtering but `GetSummary` fails, `RecallRecent` returns an error wrapping the backend failure and naming the affected session key. The parent context adapter already treats recall-provider errors as non-fatal and logs them.

### Preserve Filtering Order

The adapter continues to:

1. Search with `topN * 2`.
2. Exclude the current session.
3. Apply `minRank`.
4. Load summaries only for candidates that survive filtering.
5. Stop after `topN` matches.

This avoids introducing unnecessary summary calls for filtered-out rows.

## Risks

- One bad summary row now suppresses all recall matches for that turn. This is preferable to silently injecting incomplete context because the recall layer cannot guarantee that returned matches are meaningful without summaries.
