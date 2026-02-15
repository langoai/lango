## 1. Provider Constructor Signature Changes

- [x] 1.1 Update `anthropic.NewProvider` to accept `id string` as first parameter and use it instead of hardcoded `"anthropic"`
- [x] 1.2 Update `gemini.NewProvider` to accept `id string` parameter (after `ctx`) and use it instead of hardcoded `"gemini"`

## 2. Supervisor Call Site Updates

- [x] 2.1 Update Anthropic constructor call in `supervisor.initializeProviders` to pass `id` from config key
- [x] 2.2 Update Gemini constructor call in `supervisor.initializeProviders` to pass `id` from config key

## 3. Test Updates

- [x] 3.1 Update `TestNewProvider` in `anthropic_test.go` to pass custom ID and verify it propagates
- [x] 3.2 Update `TestAnthropicProvider_ListModels` in `anthropic_test.go` to use new constructor signature

## 4. Verification

- [x] 4.1 Run `go build ./...` to verify no compilation errors
- [x] 4.2 Run `go test ./...` to verify all tests pass
