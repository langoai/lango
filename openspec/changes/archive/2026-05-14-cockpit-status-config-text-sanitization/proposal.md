## Why

The cockpit Status page already sanitizes runtime-fed feature and graph-admission metadata, but the config-fed provider and model labels in the System section still render raw. Malformed config values can therefore leak control sequences or embedded newlines into an operator-facing status surface.

## What Changes

- Sanitize rendered provider and model values in the Status page System section.
- Add regression coverage for malformed config-fed system labels.
- Record the expanded Status page text-sanitization contract in OpenSpec and downstream docs.

## Impact

- Aligns Status page system labels with the same plain-text baseline already enforced for other status metadata.
- Removes the remaining raw config-text leak from the cockpit Status page.
