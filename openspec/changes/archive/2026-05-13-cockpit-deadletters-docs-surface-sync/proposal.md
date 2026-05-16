## Why

The cockpit feature reference already documents Mission Control, Approvals, Sessions, and Tasks in detail, but Dead Letters still appears mostly as a roster row. That leaves a high-value operator surface under-documented even though the runtime supports filters, detail inspection, retry requests, and explicit degraded/failure states.

## What Changes

- Add a dedicated Dead Letters section to the cockpit feature reference.
- Describe the filter controls, retry flow, empty/unavailable/load-failure states, and follow-up messaging.
- Extend downstream docs-sync requirements so this richer Dead Letters surface stays documented.

## Impact

- Public cockpit docs match the current Dead Letters operator surface much more closely.
- Future truth drift on Dead Letters docs becomes easier to catch.
