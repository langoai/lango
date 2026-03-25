## Delta: agent-error-handling

### Added Requirements

- **REQ-ERR-FUNCRESP-SPLIT**: `convertMessages()` MUST split merged Content with 2+ FunctionResponse parts into individual `provider.Message` entries, each with its own `tool_call_id`. This prevents OpenAI API `400: No tool output found` errors caused by EventsAdapter's consecutive same-role merge.
  - Scope: Only triggers when normalized role is `"tool"` and FunctionResponse count >= 2.

- **REQ-ERR-DANGLING-AUTHOR**: `closeDanglingParentToolCalls()` MUST set `Author` on synthetic tool-response messages to the originating assistant's Author (not hardcoded `"tool"`). Fallback order: `OriginAuthor` → `rootAgentName` → `"lango-agent"`. A warning MUST be logged when OriginAuthor is empty.

- **REQ-ERR-ORPHAN-MESSAGE**: `repairOrphanedToolCalls()` synthetic error content MUST describe the interruption cause accurately (not claim "timeout") and instruct the model not to retry the same call.

- **REQ-ERR-TUI-STDLIB-REDIRECT**: TUI mode MUST redirect Go stdlib `log` package output to the chat log file, preventing third-party library log messages (e.g., ADK runner) from corrupting the TUI display.

### Modified Behavior

- `buildToolResponseMessage` helper added for individual FunctionResponse → provider.Message conversion
- `danglingCall` struct replaces `internal.ToolCall` in `danglingToolCalls()` return type
- `closeDanglingParentToolCalls` diagnostic logging: count, origin_authors, call_ids (no payload)
- `cmd/lango/main.go:runChat` adds `log.SetOutput(logFile)` after `logging.Init()`
