## 1. Project Setup

- [x] 1.1 Initialize Go module with `go mod init github.com/langoai/lango`
- [x] 1.2 Create project directory structure (`cmd/`, `internal/`, `pkg/`)
- [x] 1.3 Add core dependencies to go.mod (ADK-Go, zap, chi, gorilla/websocket)
- [x] 1.4 Create Makefile with build/test/lint targets
- [x] 1.5 Setup cross-compilation for darwin/linux

## 2. Configuration System

- [x] 2.1 Define config struct in `internal/config/types.go`
- [x] 2.2 Implement JSON/YAML loader with viper in `internal/config/loader.go`
- [x] 2.3 Implement environment variable substitution
- [x] 2.4 Add config validation with error messages
- [x] 2.5 Write unit tests for config loading

## 3. Logging Infrastructure

- [x] 3.1 Setup zap logger with JSON/console output in `internal/logging/`
- [x] 3.2 Create subsystem logger factory (agent, gateway, channels)
- [x] 3.3 Add log level configuration from config file

## 4. Session Store

- [x] 4.1 Define session interface in `internal/session/store.go`
- [x] 4.2 Implement SQLite store in `internal/session/sqlite.go`
- [x] 4.3 Create session model with history, metadata fields
- [x] 4.4 Implement CRUD operations (Create, Get, Update, Delete)
- [x] 4.5 Add session expiration/cleanup goroutine
- [x] 4.6 Write integration tests with test database

## 5. Agent Runtime

- [x] 5.1 Setup ADK-Go agent wrapper in `internal/agent/runtime.go`
- [x] 5.2 Implement tool registration interface
- [x] 5.3 Create streaming response handler
- [x] 5.4 Integrate session store for context management
- [x] 5.5 Add multi-model provider support (Anthropic, OpenAI, Google)
- [x] 5.6 Implement context window overflow handling

## 6. Tool - Exec

- [x] 6.1 Create exec tool in `internal/tools/exec/exec.go`
- [x] 6.2 Implement synchronous command execution with timeout
- [x] 6.3 Add PTY support using creack/pty
- [x] 6.4 Implement background process management with session IDs
- [x] 6.5 Add environment variable filtering for security
- [x] 6.6 Write tests for timeout and PTY scenarios

## 7. Tool - Filesystem

- [x] 7.1 Create filesystem tool in `internal/tools/filesystem/`
- [x] 7.2 Implement file read with encoding detection
- [x] 7.3 Implement file write with atomic write support
- [x] 7.4 Implement file edit with line range replacement
- [x] 7.5 Add directory listing with metadata
- [x] 7.6 Add path traversal protection
- [x] 7.7 Write tests for edge cases

## 8. Tool - Browser

- [x] 8.1 Create browser tool in `internal/tools/browser/`
- [x] 8.2 Setup rod browser launcher and session manager
- [x] 8.3 Implement page navigation with wait states
- [x] 8.4 Implement screenshot capture (full page, element)
- [x] 8.5 Add DOM click/type/getText operations
- [x] 8.6 Add JavaScript execution support
- [x] 8.7 Write integration tests with headless browser

## 9. Gateway Server

- [x] 9.1 Create gateway package in `internal/gateway/`
- [x] 9.2 Implement HTTP server with chi router
- [x] 9.3 Implement WebSocket upgrade and connection pool
- [x] 9.4 Add RPC method handler with JSON dispatch
- [x] 9.5 Implement event broadcasting to clients
- [x] 9.6 Add health check and status endpoints
- [x] 9.7 Write integration tests for WebSocket flow

## 10. Channel - Telegram

- [x] 10.1 Create Telegram channel in `internal/channels/telegram/`
- [x] 10.2 Implement bot connection with telegram-bot-api
- [x] 10.3 Add message handler with agent forwarding
- [x] 10.4 Implement response sender with message chunking
- [x] 10.5 Add media download handling
- [x] 10.6 Implement allowlist filtering
- [x] 10.7 Write tests with mock Telegram API

## 11. Channel - Discord

- [x] 11.1 Create Discord channel in `internal/channels/discord/`
- [x] 11.2 Implement bot connection with discordgo
- [x] 11.3 Add message handler for DMs and mentions
- [x] 11.4 Implement slash command registration
- [x] 11.5 Add response sender with Discord markdown
- [x] 11.6 Write tests with mock Discord gateway

## 12. Channel - Slack

- [x] 12.1 Create Slack channel in `internal/channels/slack/`
- [x] 12.2 Implement Socket Mode connection with slack-go
- [x] 12.3 Add event handler for app_mention and message events
- [x] 12.4 Implement response sender with Block Kit
- [x] 12.5 Add thread reply support
- [x] 12.6 Write tests with mock Slack events

## 13. CLI

- [x] 13.1 Create CLI entry point in `cmd/lango/main.go`
- [x] 13.2 Add `serve` command to start gateway
- [x] 13.3 Add `version` command
- [x] 13.4 Add `config validate` command
- [x] 13.5 Implement graceful shutdown with signal handling

## 14. Integration & Testing

- [x] 14.1 Verify all tests pass
- [x] 14.2 Verify build succeeds
- [x] 14.3 Setup GitHub Actions CI with lint/test/build
- [x] 14.4 Create Docker build for multi-platform images

## 15. Documentation

- [x] 15.1 Write README.md with quick start guide
- [x] 15.2 Document configuration options
- [x] 15.3 Add channel setup guides (Telegram, Discord, Slack)

