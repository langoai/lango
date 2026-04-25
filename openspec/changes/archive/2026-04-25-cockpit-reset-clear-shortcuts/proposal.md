## Why

The cockpit dead-letter filter bar now carries enough state that operators need a single shortcut to return to the default backlog view without manually clearing every field and toggle.

## What Changes

- add a `Ctrl+R` full filter reset shortcut to the cockpit dead-letter page
- reset all draft and applied filter state back to defaults
- clear retry confirm state during reset
- ignore reset while retry is running
- document the landed reset/clear-shortcuts slice in public docs and main OpenSpec specs

## Impact

- faster operator recovery from over-filtered backlog states
- no new backend endpoints or bridge contracts
- no change to retry execution semantics
