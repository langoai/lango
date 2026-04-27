## Why

The first cockpit summary strip already shows global dead-letter counts, but operators still need faster visibility into the dominant current failure reasons without leaving the dead-letters page.

## What Changes

- extend the page-top cockpit summary strip
- add top 5 latest dead-letter reasons in a compact second `reasons:` line
- recompute that richer strip from the currently loaded backlog rows
- document the richer cockpit summary slice in public docs and main OpenSpec specs

## Impact

- richer cockpit-native dead-letter summary without a new backend summary service
- no new page or pane
- no new control-plane behavior
