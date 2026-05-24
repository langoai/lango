## Design Summary

This workstream adds the first grouped summary axis for dead-letter latest reasons.

Shared reason-family taxonomy:

- `retry-exhausted`
- `policy-blocked`
- `receipt-invalid`
- `background-failed`
- `unknown`

Classification rules:

- aggregate from each backlog row's current `latest_dead_letter_reason`
- match with case-insensitive built-in heuristics
- fall back to `unknown` when no family matches
- keep raw top latest dead-letter reasons available

CLI behavior:

- `lango status dead-letter-summary --output json` includes `by_reason_family`
- table output includes a `By reason family` section
- `top_latest_dead_letter_reasons` remains available as the raw latest reason-string view

Cockpit behavior:

- the dead-letters page summary strip includes a compact `reason families:` line
- the existing `reasons:`, `actors:`, and `dispatch:` strip lines remain available

This slice is additive. It does not add actor families, dispatch families, configurable taxonomy, persisted reason normalization, trend windows, backend summary services, or new recovery behavior.
