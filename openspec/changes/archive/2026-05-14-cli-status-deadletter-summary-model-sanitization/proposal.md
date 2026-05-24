## Why

`lango status dead-letter-summary --output json` still preserves raw bucket and top-item labels from backlog rows. That leaves summary JSON payloads vulnerable to ANSI/OSC control text or embedded newlines even though the rendered table view is already sanitized.

## What Changes

- sanitize dead-letter summary labels before they enter the aggregated summary model
- add regression coverage for malformed summary JSON labels
- sync the CLI status spec and public docs with the replay-safe summary model contract

## Impact

- keeps dead-letter summary JSON output stable for operators and downstream automation
- aligns the summary model with the existing plain single-line status rendering baseline
