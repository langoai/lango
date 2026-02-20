## 1. Fix AppendEvent - FunctionCall/FunctionResponse preservation

- [x] 1.1 Preserve original FunctionCall.ID in AppendEvent, falling back to synthetic ID only when empty
- [x] 1.2 Store FunctionResponse metadata (ID, Name, Output) in ToolCalls array alongside Content

## 2. Fix EventsAdapter - FunctionCall/FunctionResponse reconstruction

- [x] 2.1 Reconstruct FunctionCall parts with stored ID in assistant/model messages
- [x] 2.2 Reconstruct FunctionResponse parts from new-format tool messages with ToolCalls Output
- [x] 2.3 Implement legacy fallback for tool messages without ToolCalls using positional matching
- [x] 2.4 Add sequence safety to tokenBudgetTruncate for truncated history boundaries

## 3. Fix convertMessages - FunctionCall.ID forwarding

- [x] 3.1 Forward original FunctionCall.ID in convertMessages, with synthetic fallback

## 4. Tests

- [x] 4.1 Add tests for FunctionCall.ID preservation and fallback in AppendEvent
- [x] 4.2 Add tests for FunctionResponse metadata storage in AppendEvent
- [x] 4.3 Add tests for FunctionResponse reconstruction (new format and legacy fallback)
- [x] 4.4 Add tests for truncation sequence safety
