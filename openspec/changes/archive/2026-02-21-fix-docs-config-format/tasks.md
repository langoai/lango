## 1. Core Configuration Documentation

- [x] 1.1 Convert all YAML config blocks in `docs/configuration.md` to JSON (19 blocks)
- [x] 1.2 Add intro paragraph explaining config is managed via TUI or JSON import

## 2. Feature Documentation

- [x] 2.1 Convert YAML blocks in `docs/features/embedding-rag.md` (5 blocks)
- [x] 2.2 Convert YAML block in `docs/features/system-prompts.md` (1 block)
- [x] 2.3 Convert YAML block in `docs/features/knowledge-graph.md` (1 block)
- [x] 2.4 Convert YAML block in `docs/features/observational-memory.md` (1 block)
- [x] 2.5 Convert YAML block in `docs/features/knowledge.md` (1 block)
- [x] 2.6 Convert YAML block in `docs/features/librarian.md` (1 block)
- [x] 2.7 Convert YAML block in `docs/features/a2a-protocol.md` (1 block)
- [x] 2.8 Convert YAML block in `docs/features/multi-agent.md` (1 block)
- [x] 2.9 Convert YAML block in `docs/features/skills.md` (1 block)
- [x] 2.10 Convert YAML blocks in `docs/features/ai-providers.md` (5 blocks)
- [x] 2.11 Convert YAML blocks in `docs/features/channels.md` (4 blocks)

## 3. Security Documentation

- [x] 3.1 Convert YAML blocks in `docs/security/authentication.md` (3 blocks)
- [x] 3.2 Convert YAML blocks in `docs/security/encryption.md` (4 blocks)
- [x] 3.3 Convert YAML blocks in `docs/security/tool-approval.md` (7 blocks)
- [x] 3.4 Convert YAML blocks in `docs/security/pii-redaction.md` (4 blocks)
- [x] 3.5 Convert YAML block in `docs/security/index.md` (1 block)

## 4. Automation Documentation

- [x] 4.1 Convert YAML config block in `docs/automation/workflows.md` (1 config block, keep DAG YAML)
- [x] 4.2 Convert YAML block in `docs/automation/cron.md` (1 block)
- [x] 4.3 Convert YAML block in `docs/automation/background.md` (1 block)
- [x] 4.4 Convert YAML block in `docs/automation/index.md` (1 block)

## 5. Gateway Documentation

- [x] 5.1 Convert YAML block in `docs/gateway/http-api.md` (1 block)
- [x] 5.2 Convert YAML block in `docs/gateway/websocket.md` (1 block)

## 6. Payment Documentation

- [x] 6.1 Convert YAML block in `docs/payments/x402.md` (1 block)
- [x] 6.2 Convert YAML block in `docs/payments/usdc.md` (1 block)

## 7. Verification

- [x] 7.1 Grep for remaining `yaml` config blocks (only Docker Compose and workflow DAGs should remain)
- [x] 7.2 Run `go build ./...` to verify no build impact
