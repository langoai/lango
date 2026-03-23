## 1. Browser Output Signals

- [x] 1.1 Add request-local browser search state for churn diagnostics
- [x] 1.2 Enrich `browser_navigate`, `browser_search`, and `browser_extract` outputs with page-type, result-count, and empty-state signals
- [x] 1.3 Add browser unit tests for new output shape and churn state

## 2. Prompt Convergence Rules

- [x] 2.1 Update browser guidance in shared prompts to enforce extract-first bounded search behavior
- [x] 2.2 Update navigator-specific prompts and embedded AGENT definition with the bounded search workflow
- [x] 2.3 Update downstream docs/README to reflect the bounded search behavior

## 3. Verification

- [x] 3.1 Update main OpenSpec specs for browser output and prompt convergence rules
- [x] 3.2 Run `go build ./...` and `go test ./...`
- [x] 3.3 Archive the change
