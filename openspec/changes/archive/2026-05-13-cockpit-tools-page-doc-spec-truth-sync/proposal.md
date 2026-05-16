## Why

The public cockpit feature page still describes the Tools page as a tool inventory with agent assignments and invocation counts. That is not what the implementation renders today.

`internal/cli/cockpit/pages/tools.go` implements a read-only catalog browser: categories on the left, selected-category tool details on the right, with up/down cursor navigation. The spec for the page also still talks about Enter-based selection and enabled badges that do not exist in the rendered surface.

## What Changes

- Update the public cockpit feature page to describe the Tools page as a read-only category-and-tool catalog browser.
- Sync the cockpit tools-page spec to match the real navigation and rendering contract.

## Impact

- Cockpit documentation stops overstating unavailable tool-page features.
- The cockpit tools-page spec once again matches the real Bubble Tea implementation and tests.
