## Why

The cockpit documentation says the Status and Settings pages are core surfaces, but startup wiring still registers them only when specific optional dependencies are present. Both page constructors already degrade safely when those dependencies are absent.

That means the runtime can unnecessarily hide first-class cockpit surfaces even though the page implementations can render useful degraded states.

## What Changes

- Always register the cockpit Status page, even when feature-status or metrics dependencies are absent.
- Always register the cockpit Settings page, even when no config-profile store is available.
- Extract page registration into a helper and add wiring regression coverage.

## Impact

- Cockpit core navigation becomes more stable and predictable.
- Public cockpit docs about Status and Settings being always available become true in runtime behavior.
