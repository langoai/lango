## Why

The cockpit dead-letter summary strip already shows backlog overview and top latest dead-letter reasons, but operators still need faster visibility into who most recently drove manual replay activity without leaving the dead-letters page.

## What Changes

- extend the page-top cockpit summary strip
- add top 5 latest manual replay actors in a compact third `actors:` line
- recompute that richer strip from the currently loaded backlog rows
- document the actor-rich cockpit summary slice in public docs and main OpenSpec specs

## Impact

- richer cockpit-native dead-letter summary without a new backend summary service
- no new page or pane
- no new control-plane behavior
