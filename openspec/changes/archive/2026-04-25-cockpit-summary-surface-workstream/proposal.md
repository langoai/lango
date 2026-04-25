## Why

The cockpit dead-letter page already supports filtering, selection, detail inspection, and retry, but it still lacks an immediate backlog overview inside the cockpit itself.

## What Changes

- add a page-top compact summary strip to the dead-letters cockpit page
- aggregate the summary directly from the already-loaded backlog rows
- show total dead letters, retryable count, adjudication distribution, and latest-family distribution
- document the landed cockpit summary strip in public docs and main OpenSpec specs

## Impact

- first cockpit-native summary surface for dead letters
- no new backend summary service
- no new page or pane
