## Why

Most chat transcript surfaces now sanitize terminal control sequences before display, but user transcript blocks still store submitted text with only `TrimSpace`. A pasted prompt containing ANSI/OSC escape sequences can still leak raw control text into the `You` transcript block.

## What Changes

- Strip terminal control sequences from displayed user transcript content before storing it in the transcript view.
- Add regression coverage for escaped user transcript content.
- Record the user-block sanitization contract in OpenSpec and downstream cockpit docs.

## Impact

- Extends the plain visible-text baseline to user transcript blocks.
- Prevents pasted terminal control sequences from destabilizing the user side of the transcript.
