## Why

The `lango status` dashboard and dead-letter subcommands render provider/model labels, feature names/details, channels, reasons, actors, dispatch references, and other operator-facing text directly from config or runtime data. Those strings currently pass through without normalization, so malformed values can leak control sequences or embedded newlines into a primary production CLI surface.

## What Changes

- Sanitize display-facing text in the status dashboard and dead-letter CLI renderers.
- Add regression coverage for malformed status/dead-letter labels.
- Record the plain-text status CLI contract in OpenSpec and downstream docs.

## Impact

- Brings the production status CLI up to the same plain-text baseline already enforced across the cockpit surfaces.
- Prevents malformed runtime/config text from destabilizing the operator’s main diagnostic commands.
