## 1. Provider Fixes

- [x] 1.1 Pass `OutputDimensionality` in Google provider's `EmbedContent` call (`internal/embedding/google.go`)
- [x] 1.2 Add `Dimensions` field to OpenAI provider's `EmbeddingRequest` (`internal/embedding/openai.go`)
- [x] 1.3 Add `Dimensions` field to Local provider's `EmbeddingRequest` (`internal/embedding/local.go`)

## 2. Verification

- [x] 2.1 Run `go build ./...` to verify successful compilation
- [x] 2.2 Run `go test ./internal/embedding/...` to verify existing tests pass
