## 1. Prompt Update

- [x] 1.1 Add strict Markdown formatting conventions rule to `prompts/CONVERSATION_RULES.md` with sub-rules for dash-style bullets, backtick-wrapped identifiers, paired markers, and closed code blocks

## 2. Verification

- [x] 2.1 Verify `go build ./...` succeeds with updated embedded prompt file
- [x] 2.2 Verify the conversation rules section renders correctly via `DefaultBuilder().Build()` output
