## Why

Mission Control is documented as durable-first: a top-level composer submit should create a durable mission row before the shared turn is dispatched. But when `missionService` is absent, the current implementation silently falls back to generic chat execution for ordinary non-slash submits.

That breaks the durable-first contract and hides an unavailable core dependency behind a successful-looking fallback path.

## What Changes

- Keep slash-command submits working without mission service.
- Make ordinary Mission Control composer submits fail closed with an explicit system message when `missionService` is absent.
- Add a regression that ensures the shared chat executor is not called on that path.
- Sync cockpit docs and cockpit-pages spec.

## Impact

- Mission Control no longer bypasses mission creation silently when its durable mission service is missing.
- Operators get an explicit explanation instead of an invisible contract violation.
