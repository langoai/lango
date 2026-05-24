## Why

Channel transcript rows already sanitize remote message text, but the remote sender name is still rendered without stripping terminal escape sequences first. That leaves an avoidable terminal-control injection path through external channel display names.

## What Changes

- Sanitize remote channel sender names with the same ANSI/OSC stripping used for message text.
- Add regression coverage for sender-name escape stripping.
- Record the remote sender/text sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Closes a remaining external-input rendering risk in the cockpit chat transcript.
- Keeps sender-name rendering aligned with the existing hardened channel-message path.
