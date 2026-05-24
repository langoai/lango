## Why

The richer cockpit summary strip already shows top latest dead-letter reasons and top latest manual replay actors, but operators still need quick visibility into which dispatch references dominate the current backlog.

## What Changes

- extend the page-top cockpit summary strip
- add top 5 latest dispatch references in a compact `dispatch:` line
- recompute that richer strip from the currently loaded backlog rows
- document the dispatch-rich cockpit summary slice in public docs and main OpenSpec specs

## Impact

- richer cockpit-native dead-letter summary without a new backend summary service
- no new pane or page
- no new control-plane behavior
