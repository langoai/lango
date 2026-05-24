## Why

The `lango status` table renderers now sanitize operator-facing labels, but the underlying `StatusInfo` model still stores provider, model, feature detail, channel, and live feature-reason text in raw form. That leaves JSON output and any downstream reuse dependent on renderer-level sanitization instead of the collected status model itself being display-safe.

## What Changes

- Sanitize display-facing `StatusInfo` and `FeatureInfo` fields at collection/enrichment time.
- Add regression coverage for malformed collected status model text.
- Record the status-model replay-safety contract in OpenSpec and downstream docs.

## Impact

- Extends status hardening beyond the table renderer into the shared status data model and JSON path.
- Prevents raw control text from persisting inside collected status snapshots.
