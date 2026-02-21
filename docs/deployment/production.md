---
title: Production Checklist
---

# Production Checklist

Use this checklist before deploying Lango to production.

## Blockchain & Payments

- [ ] Switch to mainnet chain ID (`payment.network.chainId: 8453` for Base)
- [ ] Set appropriate per-transaction limit (`payment.limits.maxPerTx`)
- [ ] Set appropriate daily limit (`payment.limits.maxDaily`)
- [ ] Use RPC signer for production wallet security (`payment.walletProvider: rpc`)

## Security

- [ ] Enable PII redaction (`security.interceptor.redactPii: true`)
- [ ] Configure OIDC authentication for the HTTP gateway
- [ ] Enable tool approval policies for dangerous tools
- [ ] Review browser tool settings and restrict as needed

## Server & Networking

- [ ] Configure allowed origins for CORS (`server.allowedOrigins`)
- [ ] Set up structured logging (`logging.format: json`)
- [ ] Set appropriate session TTL (`session.ttl`)
- [ ] Set appropriate max history turns for memory management

## Monitoring

- [ ] Set up health endpoint monitoring (`GET /health`)
- [ ] Configure Docker health check in production compose file
- [ ] Verify logging output is captured by your log aggregator

## Automation

- [ ] Configure cron delivery channels for scheduled task output
- [ ] Review and test workflow definitions before enabling

## Verification

After deployment, verify each item:

```bash
# Health check
curl -f http://your-host:18789/health

# Docker health status
docker inspect --format='{{.State.Health.Status}}' lango
```

## Related

- [Docker](docker.md) -- Container setup and configuration
- [USDC Payments](../payments/usdc.md) -- Payment configuration reference
- [Configuration](../getting-started/configuration.md) -- Full configuration reference
