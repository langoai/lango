## 1. Fix Healthcheck

- [x] 1.1 Replace `curl -sf` with `wget --no-verbose --tries=1 --spider` in docker-compose.yml Chrome sidecar healthcheck (line 70)

## 2. Verification

- [x] 2.1 Run `docker compose --profile browser-sidecar up -d` and verify Chrome container reaches healthy status
