# Automation

Lango provides three automation systems for scheduling, background execution, and multi-step workflow orchestration. Each system is independently configurable and can be enabled or disabled via configuration flags.

## System Comparison

| Feature | [Cron](cron.md) | [Background](background.md) | [Workflow](workflows.md) |
|---------|------|------------|----------|
| Schedule Type | Cron / interval / one-time | On-demand | DAG-based YAML |
| Persistence | Ent ORM (survives restarts) | In-memory (ephemeral) | Ent ORM (survives restarts) |
| Concurrency | Configurable max jobs | Semaphore-controlled | Parallel step execution |
| Delivery | Multi-channel | Origin channel | Multi-channel + per-step |
| Session Mode | Isolated or shared | Always isolated | Per-step isolated |
| Status | **Stable** | **Experimental** | **Experimental** |

## Enabling Automation

Each system is disabled by default. Enable them in your `config.yaml`:

```yaml
cron:
  enabled: true

background:
  enabled: true

workflow:
  enabled: true
```

## Common Patterns

All three systems share the same `AgentRunner` interface to execute agent prompts, avoiding import cycles between the automation packages and the application layer:

```go
type AgentRunner interface {
    Run(ctx context.Context, sessionKey string, prompt string) (string, error)
}
```

Results can be delivered to communication channels (Telegram, Discord, Slack) via the `ChannelSender` / `Delivery` adapters. Each system supports a `defaultDeliverTo` configuration option as a fallback when no delivery target is specified per-job or per-workflow.
