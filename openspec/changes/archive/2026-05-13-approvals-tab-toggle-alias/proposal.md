## Why

The cockpit Approvals page currently requires `/` to switch between History and Grants. That works, but it is not an intuitive default for a two-section TUI surface, and the existing tests still reflect a lingering expectation that `Tab` should switch sections.

## What Changes

- Let the Approvals page accept `Tab` as an additional section-toggle key while keeping `/` as a compatibility alias.
- Update the help text so operators can discover the more intuitive `Tab` key.
- Lock the behavior with regressions and sync the approvals cockpit spec.

## Impact

- Approvals navigation becomes more intuitive without breaking existing `/` muscle memory.
- Tests, help, and spec all describe the same section-toggle contract.
