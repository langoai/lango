## 1. Dockerfile Improvements

- [x] 1.1 Add non-root user (useradd, USER directive)
- [x] 1.2 Add health check (call /health endpoint with curl or wget)
- [x] 1.3 Set data directory permissions (chown)

## 2. docker-compose.yml Creation

- [x] 2.1 Define lango service (build, ports, volumes)
- [x] 2.2 Set environment variables (ANTHROPIC_API_KEY, DISCORD_BOT_TOKEN, TELEGRAM_BOT_TOKEN, SLACK_BOT_TOKEN, SLACK_APP_TOKEN)
- [x] 2.3 Define volumes (lango-data)
- [x] 2.4 Set lango.json mount (read-only)

## 3. Docker Environment Detection Logic

- [x] 3.1 Add Docker environment detection function (check /.dockerenv or /proc/1/cgroup)
- [x] 3.2 Check Docker environment during LocalCryptoProvider initialization
- [x] 3.3 Output error message when using LocalCryptoProvider in Docker environment

## 4. Verification

- [x] 4.1 Docker image build test
- [x] 4.2 docker-compose up execution confirmation
- [x] 4.3 Health check operation confirmation
- [x] 4.4 Verify RPC Provider required error message in Docker environment
