## Context

Observed production logs showed three distinct approval hashes for what users experienced as the same web search:

- `{"query":"Trump latest news"}`
- `{"query":"Trump latest news","limit":3}`
- `{"query":"Trump latest news","limit":5}`

The current replay guard treated each of those as unrelated, while timeout entries remained negative and sticky for their exact raw param shape.

## Decisions

### D1. Canonical approval identity for `browser_search`

`browser_search` approval replay identity will normalize to:

```json
{"query":"<trimmed-and-collapsed-query>"}
```

`limit` is omitted because it does not change the approval risk. The approval summary already reflects query-only semantics.

### D2. Timeout is retryable, but bounded

Timeout is different from deny:

- `deny` means the user explicitly said no, so the turn-local replay must block immediately.
- `timeout` can be accidental, so the runtime will allow a bounded number of re-prompts for the same canonical action.

The middleware will track timeout count in turn-local approval state. Once the count reaches `MaxTurnApprovalTimeouts`, later retries are replay-blocked for the rest of the turn.

### D3. Later approval supersedes earlier timeout

Because all browser-search limit variants now share one canonical key, a later approval overwrites the earlier timeout state for that semantic action. Subsequent retries bypass approval instead of reviving the stale timeout.

## Risks

- Over-canonicalizing approval keys could merge actions that should stay distinct.
  - Mitigation: limit the new normalization to `browser_search`, where `limit` is extraction-only and approval-neutral.
- Allowing timeout re-prompts could reintroduce prompt churn.
  - Mitigation: keep the retry budget bounded and continue replay-blocking deny/unavailable immediately.
