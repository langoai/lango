## Why

Mission Control rendering and snapshot models are now largely sanitized, but accepting a prepared proposal still builds the durable mission `Description` from raw brief and summary text. That means malformed proposal content can persist into durable mission storage even though the transient UI surfaces are hardened.

## What Changes

- Sanitize prepared-brief and fallback proposal description text before passing it into mission acceptance.
- Add regression coverage for malformed proposal brief persistence.
- Record the durable proposal-description sanitization contract in OpenSpec and downstream docs.

## Impact

- Extends cockpit text hardening from UI replay paths into the durable mission-persistence boundary.
- Prevents raw control text from being written into mission descriptions via proposal acceptance.
