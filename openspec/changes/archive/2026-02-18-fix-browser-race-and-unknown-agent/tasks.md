## 1. Browser Race Condition Fix

- [x] 1.1 Add `browserOnce sync.Once` and `browserErr error` fields to `browser.Tool` struct
- [x] 1.2 Refactor `ensureBrowser()` to use `sync.Once` with retry-on-failure (reset Once on error)
- [x] 1.3 Extract browser initialization logic into `initBrowser()` — assign `t.browser` only after successful `Connect()`
- [x] 1.4 Reset `browserOnce` and `browserErr` in `Close()` to allow re-initialization

## 2. Unknown Agent "model" Fix

- [x] 2.1 Add `"model"` to the `"assistant"` case in EventsAdapter role-to-author switch
- [x] 2.2 Change `default` fallback to use `rootAgentName` (or `"lango-agent"`) instead of literal `"model"`

## 3. Verification

- [x] 3.1 Run `go build ./...` and verify success
- [x] 3.2 Run `go test ./internal/tools/browser/... ./internal/adk/...` and verify all pass
