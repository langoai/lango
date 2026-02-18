## MODIFIED Requirements

### Requirement: Docker Compose Orchestration
The system SHALL provide a docker-compose.yml with deployment profiles for different browser configurations. The README documentation SHALL describe the import→delete configuration pattern instead of read-only JSON mounting.

#### Scenario: Sidecar browser deployment (browser-sidecar profile)
- **WHEN** running `docker compose --profile browser-sidecar up`
- **THEN** the lango-sidecar service SHALL start without Chromium
- **AND** a separate `chromedp/headless-shell` container SHALL run alongside
- **AND** lango SHALL connect to Chrome via `ROD_BROWSER_WS=ws://chrome:9222`
- **AND** the Chrome container SHALL have a memory limit of 512MB
- **AND** the Chrome container SHALL have a healthcheck using `wget` to verify CDP availability at `http://localhost:9222/json/version`
- **AND** the healthcheck SHALL use `wget --no-verbose --tries=1 --spider` flags
- **AND** the lango-sidecar service SHALL depend on the Chrome container being healthy (`service_healthy`)
