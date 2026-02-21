## 1. Type Change

- [x] 1.1 Change `LearningStats.ByCategory` field type from `map[string]int` to `map[entlearning.Category]int`
- [x] 1.2 Update `make()` call in `GetLearningStats` to `make(map[entlearning.Category]int)`
- [x] 1.3 Remove `string(e.Category)` cast — use `e.Category` directly as map key

## 2. Verification

- [x] 2.1 Run `go build ./...` — zero errors
- [x] 2.2 Run `go test ./internal/knowledge/...` — all pass
