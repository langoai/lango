## Why

The Status page runtime now shows explicit unavailable-state copy when the feature-status provider or observability collector is missing. The public cockpit feature page still says only that "placeholder text is shown", which underspecifies the current behavior.

## What Changes

- Update the cockpit feature doc to describe the concrete unavailable-state messaging for the Status page.
- Extend downstream docs requirements so this public wording stays aligned with the current runtime contract.

## Impact

- Public cockpit docs match the current Status page behavior.
- Operators are told exactly which sections degrade when status dependencies are absent.
