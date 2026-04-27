## 1. Shared reason-family classifier

- [x] 1.1 Add a shared latest dead-letter reason-family classifier
- [x] 1.2 Cover retry-exhausted, policy-blocked, receipt-invalid, background-failed, and unknown mappings

## 2. CLI summary

- [x] 2.1 Add `by_reason_family` to the dead-letter summary result
- [x] 2.2 Aggregate reason-family buckets from latest dead-letter reasons
- [x] 2.3 Render a `By reason family` table section
- [x] 2.4 Preserve raw `top_latest_dead_letter_reasons`

## 3. Cockpit summary strip

- [x] 3.1 Aggregate reason-family buckets from current cockpit backlog rows
- [x] 3.2 Render a compact `reason families:` strip line
- [x] 3.3 Preserve raw `reasons:` strip output

## 4. Docs and OpenSpec

- [x] 4.1 Document CLI `by_reason_family` and table section
- [x] 4.2 Document cockpit `reason families:` strip line
- [x] 4.3 Document the initial reason-family taxonomy
- [x] 4.4 State that raw top latest dead-letter reasons remain available
- [x] 4.5 Sync main docs-only OpenSpec requirements
- [x] 4.6 Archive the completed workstream
