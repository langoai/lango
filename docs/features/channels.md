---
title: Channels
---

# Channels

Lango supports multi-channel communication, allowing your agent to interact with users across different messaging platforms simultaneously.

## Supported Channels

| Channel | Config Section | Implementation |
|---------|---------------|----------------|
| **Telegram** | `channels.telegram` | `internal/channels/telegram/` |
| **Discord** | `channels.discord` | `internal/channels/discord/` |
| **Slack** | `channels.slack` | `internal/channels/slack/` |

Each channel runs as an independent integration within the same Lango process. Messages from all channels are routed to the same agent, maintaining separate sessions per user/channel.

## Setup

The easiest way to configure channels is through the onboarding wizard:

```bash
lango onboard
```

Select **Channel Setup** during onboarding to configure one or more channels.

## Telegram

### Prerequisites

1. Create a bot via [BotFather](https://t.me/BotFather) on Telegram
2. Copy the bot token

### Configuration

> **Settings:** `lango settings` → Channels

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "botToken": "${TELEGRAM_BOT_TOKEN}",
      "allowlist": []
    }
  }
}
```

| Key | Type | Description |
|-----|------|-------------|
| `enabled` | `bool` | Enable the Telegram channel |
| `botToken` | `string` | Bot token from BotFather |
| `allowlist` | `[]int64` | Allowed user/group IDs (empty = allow all) |

!!! warning "Security"

    In production, always set `allowlist` to restrict which Telegram users and groups can interact with your agent.

## Discord

### Prerequisites

1. Create an application in the [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a bot user and copy the bot token
3. Note the Application ID for slash command registration

### Configuration

> **Settings:** `lango settings` → Channels

```json
{
  "channels": {
    "discord": {
      "enabled": true,
      "botToken": "${DISCORD_BOT_TOKEN}",
      "applicationId": "your-application-id",
      "allowedGuilds": []
    }
  }
}
```

| Key | Type | Description |
|-----|------|-------------|
| `enabled` | `bool` | Enable the Discord channel |
| `botToken` | `string` | Bot token from Discord Developer Portal |
| `applicationId` | `string` | Application ID for slash commands |
| `allowedGuilds` | `[]string` | Allowed guild (server) IDs (empty = allow all) |

## Slack

### Prerequisites

1. Create a Slack app at [api.slack.com](https://api.slack.com/apps)
2. Enable Socket Mode for real-time events
3. Add required bot scopes and install to your workspace

### Configuration

> **Settings:** `lango settings` → Channels

```json
{
  "channels": {
    "slack": {
      "enabled": true,
      "botToken": "${SLACK_BOT_TOKEN}",
      "appToken": "${SLACK_APP_TOKEN}",
      "signingSecret": "${SLACK_SIGNING_SECRET}"
    }
  }
}
```

| Key | Type | Description |
|-----|------|-------------|
| `enabled` | `bool` | Enable the Slack channel |
| `botToken` | `string` | Bot OAuth token (`xoxb-...`) |
| `appToken` | `string` | App-level token for Socket Mode (`xapp-...`) |
| `signingSecret` | `string` | Signing secret for request verification |

## Channel Features

All channels share the following capabilities:

- **Session isolation** -- Each user/channel combination gets its own session
- **Tool approval** -- Interactive approval prompts forwarded to the originating channel
- **Message formatting** -- Markdown/rich text adapted per platform
- **Delivery targets** -- Automation systems (cron, background, workflow) can deliver results to any enabled channel

## Multiple Channels

You can enable multiple channels simultaneously. Each runs independently:

> **Settings:** `lango settings` → Channels

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "botToken": "${TELEGRAM_BOT_TOKEN}"
    },
    "discord": {
      "enabled": true,
      "botToken": "${DISCORD_BOT_TOKEN}",
      "applicationId": "123456789"
    },
    "slack": {
      "enabled": true,
      "botToken": "${SLACK_BOT_TOKEN}",
      "appToken": "${SLACK_APP_TOKEN}",
      "signingSecret": "${SLACK_SIGNING_SECRET}"
    }
  }
}
```

## Related

- [Tool Approval](../security/tool-approval.md) -- How approval prompts work across channels
- [Cron Scheduling](../automation/cron.md) -- Deliver scheduled results to channels
- [Background Tasks](../automation/background.md) -- Deliver async results to channels
