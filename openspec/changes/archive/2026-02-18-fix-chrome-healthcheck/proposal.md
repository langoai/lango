## Why

The `lango-chrome` container fails healthcheck when started via `docker compose --profile browser-sidecar up -d`. The healthcheck uses `curl`, which is not available in the `chromedp/headless-shell:latest` Alpine-based image. This causes the container to be stuck in an unhealthy state, preventing dependent services from functioning.

## What Changes

- Replace `curl` with `wget` in the Chrome sidecar healthcheck command in `docker-compose.yml`
- `wget` is included by default in Alpine-based images, so no additional installation is needed

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `docker-deployment`: Fix healthcheck command compatibility with Alpine-based Chrome image

## Impact

- **File**: `docker-compose.yml` (line 70, healthcheck test command)
- **Runtime**: Chrome sidecar container will correctly report healthy status
- **Dependencies**: No new dependencies; `wget` is already present in the base image
