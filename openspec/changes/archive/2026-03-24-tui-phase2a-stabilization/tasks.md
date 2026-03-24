## 1. Mouse Wheel Scrolling

- [x] 1.1 Add `tea.WithMouseCellMotion()` to `tea.NewProgram()` call in `cmd/lango/main.go`

## 2. Log File Redirect

- [x] 2.1 Replace `Writer: os.Stderr` with `OutputPath: filepath.Join(cfg.DataRoot, "chat.log")` in `runChat()` logging init
- [x] 2.2 Display log file path in TUI startup banner before entering alt-screen

## 3. CPR Filter State Machine

- [x] 3.1 Add `cprState` type, constants (cprIdle/cprGotEsc/cprGotBracket/cprInParams), and `cprTimeoutMsg` type to `internal/cli/chat/chat.go`
- [x] 3.2 Add `cprDetect` and `cprBuf` fields to `ChatModel` struct
- [x] 3.3 Implement `filterCPR()` method with 4-state FSM transitions
- [x] 3.4 Implement `cprFlush()` method to replay buffered keys through normal handlers
- [x] 3.5 Integrate CPR filter into `Update()` — intercept `tea.KeyMsg` before `handleKey()`, handle `cprTimeoutMsg`

## 4. Testing

- [x] 4.1 Add test: full CPR sequence (ESC[43;84R) is discarded
- [x] 4.2 Add test: non-CPR sequence (ESC + non-bracket) flushes correctly
- [x] 4.3 Add test: partial sequence with non-digit flushes correctly
- [x] 4.4 Add test: timeout flushes buffered ESC
- [x] 4.5 Add test: R at cprGotBracket (no digits) is not treated as CPR

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./internal/cli/chat/ -v` passes all tests including new CPR tests
