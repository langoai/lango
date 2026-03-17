## 1. Critical Bug Fix — Fallback Model Leak

- [x] 1.1 Modify `internal/supervisor/proxy.go` line 99: copy `params` and reset `Model` to `""` before fallback call
- [x] 1.2 Create `internal/supervisor/proxy_test.go` with table-driven tests: primary success (no fallback), primary fail + fallback success (model reset verified), both fail (error returned), original params not mutated

## 2. Provider-Model Validation Function

- [x] 2.1 Create `internal/provider/validate.go` with `ErrModelProviderMismatch` sentinel and `ValidateModelProvider()` function using prefix blocklist
- [x] 2.2 Create `internal/provider/validate_test.go` with table-driven tests covering all provider types, empty model, case insensitivity, ollama/github passthrough

## 3. Startup Config Validation

- [x] 3.1 Modify `internal/config/loader.go` `Validate()`: add `agent.fallbackProvider` existence check against providers map
- [x] 3.2 Modify `internal/config/loader.go` `Validate()`: add primary and fallback model-provider compatibility checks using `ValidateModelProvider`

## 4. Runtime Provider Guards

- [x] 4.1 Modify `internal/provider/gemini/gemini.go` `Generate()`: add `ValidateModelProvider("gemini", model)` call after alias normalization
- [x] 4.2 Modify `internal/provider/anthropic/anthropic.go` `Generate()`: add `ValidateModelProvider("anthropic", params.Model)` call before request processing

## 5. Verification

- [x] 5.1 Run `go build ./...` — verify clean build
- [x] 5.2 Run `go test ./internal/supervisor/... -run TestProxyFallback` — all pass
- [x] 5.3 Run `go test ./internal/provider/... -run TestValidateModelProvider` — all pass
- [x] 5.4 Run `go test ./...` — full suite passes with no regressions
