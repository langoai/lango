## Why

Mission Control projected loop snapshots are largely sanitized, but the cron-loop source path still feeds raw cron job names and last-run status text into the loop projector input. That leaves one scheduled-automation path dependent on later sanitization instead of making the source snapshot itself display-safe.

## What Changes

- Sanitize cron job `Name` and `LastRunStatus` before building `loopview.CronSource`.
- Add regression coverage for malformed scheduled-loop source text.
- Record the scheduled-loop source replay-safety contract in OpenSpec and downstream docs.

## Impact

- Closes the remaining raw text path in Mission Control’s scheduled automation source input.
- Keeps projected loop rows aligned with the same plain-text baseline from source construction onward.
