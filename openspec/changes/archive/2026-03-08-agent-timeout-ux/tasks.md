## 1. Structured Errors + Partial Result Recovery (Core)

- [x] 1.1 Create `internal/adk/errors.go` with AgentError type, ErrorCode constants, classifyError
- [x] 1.2 Modify `runAndCollectOnce()` to return partial text in AgentError on failure
- [x] 1.3 Modify `RunStreaming()` to return partial text in AgentError on failure
- [x] 1.4 Modify `RunAndCollect()` to preserve best partial result through retry logic
- [x] 1.5 Add tests for AgentError, classifyError, errors.As integration

## 2. Error Formatting + Partial Result Handling (Application)

- [x] 2.1 Create `internal/app/error_format.go` with FormatUserError and formatPartialResponse
- [x] 2.2 Modify `runAgent()` in channels.go to recover partial results from AgentError
- [x] 2.3 Add formatChannelError helper with duck-typed UserMessage interface to all 3 channels
- [x] 2.4 Update channel sendError functions to use formatChannelError
- [x] 2.5 Update gateway handleChatMessage to include structured error fields in agent.error event
- [x] 2.6 Add tests for FormatUserError and formatPartialResponse

## 3. Progressive Thinking Indicators (UI)

- [x] 3.1 Add startProgressUpdates method to Slack channel with 15s periodic placeholder edit
- [x] 3.2 Wire startProgressUpdates into Slack handleMessage between postThinking and handler
- [x] 3.3 Add postThinking, editMessage, startProgressUpdates to Telegram channel
- [x] 3.4 Update Telegram handleUpdate to use thinking placeholder instead of pure typing
- [x] 3.5 Add postThinking, editPlaceholder, startProgressUpdates to Discord channel
- [x] 3.6 Update Discord onMessageCreate to use thinking placeholder instead of pure typing
- [x] 3.7 Add periodic agent.progress broadcast goroutine to gateway handleChatMessage
- [x] 3.8 Update Discord and Telegram tests to match new placeholder behavior

## 4. Auto-Extend Timeout (Enhancement)

- [x] 4.1 Add AutoExtendTimeout and MaxRequestTimeout config fields to AgentConfig
- [x] 4.2 Create `internal/app/deadline.go` with ExtendableDeadline type
- [x] 4.3 Add RunOption type and WithOnActivity callback to adk agent
- [x] 4.4 Wire onActivity callback into runAndCollectOnce and RunStreaming
- [x] 4.5 Modify runAgent to use ExtendableDeadline when AutoExtendTimeout is enabled
- [x] 4.6 Add tests for ExtendableDeadline (expiry, extension, max timeout, stop)
