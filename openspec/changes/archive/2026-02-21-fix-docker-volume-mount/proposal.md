## Why

The Docker named volume `lango-data` was mounted at `/data`, but the application stores all persistent data (lango.db, graph.db, skills/) in `~/.lango/` (`/home/lango/.lango`). This means container deletion causes complete data loss since the actual data directory was never volume-backed.

## What Changes

- Mount the `lango-data` named volume at `/home/lango/.lango` instead of `/data`
- Remove the unused `/data` directory creation from the Dockerfile
- Update the `docker-deployment` spec to reflect the correct mount path

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `docker-deployment`: Data persistence volume mount path changes from `/data` to `/home/lango/.lango` to match actual application data directory

## Impact

- `docker-compose.yml`: Volume mount path updated
- `Dockerfile`: Removed `/data` directory setup (mkdir, chown, chmod)
- Existing containers must be recreated (`docker compose down && docker compose up -d`) for the new mount to take effect
- No code changes required — only infrastructure configuration
