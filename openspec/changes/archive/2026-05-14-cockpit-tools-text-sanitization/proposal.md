## Why

The cockpit Tools page renders category names, category descriptions, tool names, tool descriptions, and safety labels directly from catalog metadata. Those values are runtime-fed and currently pass through without normalization, so malformed entries can leak control sequences or embedded newlines into a read-only operator surface. Safety color selection also keys off the raw label, so escaped known values can lose their intended styling.

## What Changes

- Sanitize operator-facing Tools page text before rendering it.
- Key safety-level styling off the sanitized safety label.
- Add regressions for malformed category and tool metadata.
- Record the plain-text Tools page contract in OpenSpec and downstream docs.

## Impact

- Brings the cockpit Tools page up to the same plain-text baseline as the hardened chat and Status surfaces.
- Prevents malformed catalog metadata from destabilizing the read-only tool browser.
