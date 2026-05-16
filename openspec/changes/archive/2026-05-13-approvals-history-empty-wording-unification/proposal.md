## Why

The Approvals page already uses `No approval history yet.` for the top-level configured-but-empty state, but the history section by itself still renders `No history entries`. That makes the same absence of approval history read differently depending on whether the grants section happens to be empty.

## What Changes

- Unify the history-section empty wording to `No approval history yet.`
- Add regression coverage for configured-empty history when the grants section is still present.
- Extend the approval-history-view spec so the section-level empty wording stays aligned.

## Impact

- The Approvals page uses one consistent history-empty message across top-level and section-level cases.
- Runtime, tests, docs, and spec describe the same empty-history contract.
