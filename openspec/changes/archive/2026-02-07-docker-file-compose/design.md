## Context

lango currently has a Dockerfile but no docker-compose.yml for production deployment. Also, the current Dockerfile does not account for passphrase handling in headless environments, making it impossible to run when using LocalCryptoProvider because no TTY is available.

**Current state:**
- Dockerfile: Basic multi-stage build exists
- docker-compose.yml: None
- Headless issue: LocalCryptoProvider requires interactive terminal

## Goals / Non-Goals

**Goals:**
- Enable stable operation of lango in Docker environments
- Enable encryption in headless environments through RPC Provider (Companion)
- Support all channels (Discord, Telegram, Slack)
- Support Browser tool (Chromium included)
- Container health monitoring via health check
- Security hardening (non-root user execution)

**Non-Goals:**
- Separate configurations for multiple environments (dev/staging/prod)
- Kubernetes/Helm chart creation
- Companion app containerization (separate project)
- Docker environment support for LocalCryptoProvider

## Decisions

### Decision 1: Forced RPC Provider Usage
In Docker environments, LocalCryptoProvider is not used; RPC Provider (Companion integration) is enforced.

**Alternatives:**
- A) Pass passphrase via environment variable → Security vulnerability (plaintext exposed in env)
- B) Mount passphrase file via Docker Secret → Complex, requires LocalCryptoProvider modification
- C) **Force RPC Provider** → Maintains existing security model, architecture separated from Companion

**Choice: Option C** - Using RPC Provider in headless environments aligns with the existing design intent.

### Decision 2: docker-compose.yml Structure

```yaml
services:
  lango:
    build: .
    ports: ["18789:18789"]
    volumes:
      - lango-data:/data
      - ./lango.json:/app/lango.json:ro
    environment:
      - ANTHROPIC_API_KEY
      - DISCORD_BOT_TOKEN
      - TELEGRAM_BOT_TOKEN
      - SLACK_BOT_TOKEN
      - SLACK_APP_TOKEN
```

**Alternatives:**
- A) All settings via environment variables → Complex, inconsistent with lango's config system
- B) **Config file mount + environment variable substitution** → Consistent with existing design

**Choice: Option B**

### Decision 3: Dockerfile Improvements

1. **Health check added**: Status verification via HTTP endpoint
2. **Non-root user**: Security hardening
3. **Chromium retained**: Required for Browser tool

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| lango cannot start without Companion connection | Clear error messages and documentation |
| Image size increase due to Chromium (~200MB) | Accepted as Browser tool requirement |
| Config file mount required | Example included in docker-compose.yml |
