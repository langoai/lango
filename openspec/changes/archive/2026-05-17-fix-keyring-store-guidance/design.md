# Design

This is a small CLI guidance fix. No architecture change is required.

The warning copy will be centralized in a small package helper used by both
passphrase-changing flows:

- `lango security change-passphrase`
- `lango security recovery restore`

Keeping the text in one helper avoids command drift while preserving existing
stdout/stderr routing.
