## Context

The `chromedp/headless-shell:latest` image is Alpine-based and does not include `curl`. The current healthcheck in `docker-compose.yml` uses `curl` to probe the Chrome DevTools Protocol endpoint, causing the container to remain unhealthy indefinitely.

## Goals / Non-Goals

**Goals:**
- Fix the Chrome sidecar healthcheck so it correctly reports healthy status
- Use a tool already available in the Alpine base image

**Non-Goals:**
- Changing the healthcheck endpoint or protocol
- Modifying the Chrome container image itself
- Adding new packages to the container

## Decisions

**Decision: Use `wget` instead of `curl`**

- `wget` is included by default in Alpine Linux via BusyBox
- The `--spider` flag performs a HEAD request without downloading content (equivalent to `curl -sf`)
- `--no-verbose` suppresses output, `--tries=1` prevents retries (healthcheck already handles retries)
- No image modification or Dockerfile changes required

Alternative considered: Install `curl` in a custom Dockerfile layer. Rejected because it adds image size, build complexity, and maintenance burden for no benefit.

## Risks / Trade-offs

- [Minimal risk] `wget --spider` sends a HEAD request vs `curl`'s GET. The CDP `/json/version` endpoint supports both methods, so this is safe.
