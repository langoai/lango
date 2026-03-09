## 1. Test Infrastructure (testutil package)

- [x] 1.1 Create `internal/testutil/helpers.go` with NopLogger, TestEntClient, SkipShort
- [x] 1.2 Create `internal/testutil/helpers_test.go` with table-driven tests
- [x] 1.3 Create `internal/testutil/mock_session_store.go` implementing session.Store
- [x] 1.4 Create `internal/testutil/mock_provider.go` implementing provider.Provider
- [x] 1.5 Create `internal/testutil/mock_embedding.go` implementing embedding.EmbeddingProvider
- [x] 1.6 Create `internal/testutil/mock_graph.go` implementing graph.Store
- [x] 1.7 Create `internal/testutil/mock_crypto.go` implementing security.CryptoProvider
- [x] 1.8 Create `internal/testutil/mock_generators.go` with MockTextGenerator, MockAgentRunner, MockChannelSender
- [x] 1.9 Create `internal/testutil/mock_cron.go` implementing cron.Store

## 2. Assertion Standardization + t.Parallel()

- [x] 2.1 Refactor `internal/adk/` test files to testify + t.Parallel()
- [x] 2.2 Refactor `internal/config/` test files to testify + t.Parallel()
- [x] 2.3 Refactor `internal/security/` test files to testify + t.Parallel()
- [x] 2.4 Refactor `internal/learning/` test files to testify + t.Parallel()
- [x] 2.5 Refactor `internal/eventbus/` test files to testify + t.Parallel()
- [x] 2.6 Refactor `internal/lifecycle/` test files to testify + t.Parallel()
- [x] 2.7 Refactor `internal/appinit/` test files to testify + t.Parallel()
- [x] 2.8 Refactor `internal/skill/` test files to testify + t.Parallel()
- [x] 2.9 Refactor `internal/toolcatalog/` test files to testify + t.Parallel()
- [x] 2.10 Refactor `internal/toolchain/` test files to testify + t.Parallel()
- [x] 2.11 Refactor `internal/tools/` test files to testify + t.Parallel()
- [x] 2.12 Refactor `internal/prompt/` test files to testify + t.Parallel()
- [x] 2.13 Refactor `internal/channels/` test files to testify + t.Parallel()
- [x] 2.14 Refactor `internal/p2p/` test files to testify + t.Parallel()
- [x] 2.15 Refactor `internal/economy/` test files to testify + t.Parallel()
- [x] 2.16 Refactor `internal/app/` test files to testify (no t.Parallel)
- [x] 2.17 Migrate `internal/gateway/` test files to testify + t.Parallel()
- [x] 2.18 Add t.Parallel() to `internal/session/child_test.go`

## 3. New Test Coverage

- [x] 3.1 Create `internal/mdparse/frontmatter_test.go` with table-driven tests
- [x] 3.2 Create `internal/logging/logger_test.go` with logger creation and level tests
- [x] 3.3 Create `internal/cron/scheduler_test.go` with lifecycle tests
- [x] 3.4 Create `internal/cron/executor_test.go` with execution tests
- [x] 3.5 Create `internal/cron/delivery_test.go` with delivery routing tests
- [x] 3.6 Create `internal/config/loader_integration_test.go` with YAML/env tests
- [x] 3.7 Create `internal/config/types_defaults_test.go` with defaults and validation tests
- [x] 3.8 Create `internal/mcp/config_loader_test.go` with config loading tests
- [x] 3.9 Create `internal/mcp/connection_test.go` with tool name formatting tests
- [x] 3.10 Create `internal/mcp/errors_test.go` with sentinel error tests
- [x] 3.11 Create `internal/mcp/adapter_test.go` with adapter function tests
- [x] 3.12 Create `internal/app/wiring_test.go` with wiring helper tests
- [x] 3.13 Create `internal/app/tools_registration_test.go` with tool registration tests
- [x] 3.14 Create `internal/app/sender_test.go` with channelSender adapter tests

## 4. Benchmarks

- [x] 4.1 Create `internal/types/token_bench_test.go` with token estimation benchmarks
- [x] 4.2 Create `internal/memory/token_bench_test.go` with message counting benchmarks
- [x] 4.3 Create `internal/prompt/builder_bench_test.go` with prompt building benchmarks
- [x] 4.4 Create `internal/graph/bolt_store_bench_test.go` with graph traversal benchmarks
- [x] 4.5 Create `internal/asyncbuf/batch_bench_test.go` with buffer operation benchmarks
- [x] 4.6 Create `internal/embedding/rag_bench_test.go` with RAG search benchmarks

## 5. Verification

- [x] 5.1 Run `go build ./...` — full project builds
- [x] 5.2 Run `go test ./internal/...` — all 89 packages pass
- [x] 5.3 Run `go vet ./internal/...` — no issues
