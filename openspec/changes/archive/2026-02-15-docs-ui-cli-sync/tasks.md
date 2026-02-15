## 1. Dockerfile Update

- [x] 1.1 Change `CGO_ENABLED=0` to `CGO_ENABLED=1` in builder stage
- [x] 1.2 Add `COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh` to runtime stage
- [x] 1.3 Change `ENTRYPOINT ["lango"]` to `ENTRYPOINT ["docker-entrypoint.sh"]`

## 2. Entrypoint Script

- [x] 2.1 Create `docker-entrypoint.sh` with passphrase keyfile setup (copy from Docker secret to `~/.lango/keyfile`, chmod 0600)
- [x] 2.2 Add first-run config import logic (copy secret to /tmp, run `lango config import`, temp file auto-deleted)
- [x] 2.3 Add idempotency check (skip import if profile already exists)
- [x] 2.4 Support configurable secret paths via `LANGO_CONFIG_FILE` and `LANGO_PASSPHRASE_FILE` env vars
- [x] 2.5 End with `exec lango "$@"` to replace entrypoint as PID 1

## 3. Docker Compose Update

- [x] 3.1 Remove `./lango.json:/app/lango.json:ro` volume mount
- [x] 3.2 Remove environment variable passthrough (ANTHROPIC_API_KEY, DISCORD_BOT_TOKEN, etc.)
- [x] 3.3 Add `secrets` block with `lango_config` (file: ./config.json) and `lango_passphrase` (file: ./passphrase.txt)
- [x] 3.4 Add `LANGO_PROFILE=default` environment variable

## 4. README Update

- [x] 4.1 Rewrite Docker Headless Configuration section to document Docker secrets workflow
- [x] 4.2 Document entrypoint script behavior (keyfile setup, first-run import, restart idempotency)
- [x] 4.3 Document optional environment variables (LANGO_PROFILE, LANGO_CONFIG_FILE, LANGO_PASSPHRASE_FILE)

## 5. Verification

- [x] 5.1 Dockerfile uses `CGO_ENABLED=1` and `ENTRYPOINT ["docker-entrypoint.sh"]`
- [x] 5.2 docker-compose.yml has no environment variable secrets, uses Docker secrets block
- [x] 5.3 docker-entrypoint.sh is executable and handles both first-run and restart scenarios
- [x] 5.4 README Docker section documents the secure secrets-based workflow
