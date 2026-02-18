## 1. Gateway Server — Filter http.ErrServerClosed

- [x] 1.1 Add `"errors"` import to `internal/gateway/server.go`
- [x] 1.2 Modify `Start()` to check `errors.Is(err, http.ErrServerClosed)` and return `nil` instead of the error

## 2. Application Stop — Downgrade cleanup error logs

- [x] 2.1 Change `Errorw("gateway shutdown error")` to `Warnw` in `internal/app/app.go` `Stop()` method
- [x] 2.2 Change `Errorw("browser close error")` to `Warnw` in `internal/app/app.go` `Stop()` method
- [x] 2.3 Change `Errorw("session store close error")` to `Warnw` in `internal/app/app.go` `Stop()` method
- [x] 2.4 Change `Errorw("graph store close error")` to `Warnw` in `internal/app/app.go` `Stop()` method

## 3. Main Shutdown Handler — Downgrade error log

- [x] 3.1 Change `Errorw("shutdown error")` to `Warnw` in `cmd/lango/main.go` shutdown goroutine

## 4. Verification

- [x] 4.1 Run `go build ./...` — confirm no compilation errors
- [x] 4.2 Run `go test ./...` — confirm all tests pass
