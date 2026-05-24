## 1. SpecDriftDetectedEvent

- [x] 1.1 Add `EventSpecDriftDetected` constant and `SpecDriftDetectedEvent` struct to `internal/eventbus/continuity_events.go`
- [x] 1.2 Add `EventName()` method implementing Event interface

## 2. EmitSpecDrift on SuggestionEmitter

- [x] 2.1 Add drift tracking state: `driftCounters map[string]int`, `driftThreshold int`
- [x] 2.2 Default threshold = 5 via `defaultDriftThreshold` constant
- [x] 2.3 Implement `EmitSpecDrift(ctx, toolName, errorClass, sampleErr) bool` with frequency tracking, dedup, and event publish
- [x] 2.4 Add unit tests: below threshold no event, threshold crossed publishes, dedup after emit

## 3. Wire into learning engine

- [x] 3.1 In `engine.go` `OnToolResult`, call `EmitSpecDrift` when err != nil with toolName and categorized error class
- [x] 3.2 Existing engine tests pass

## 4. Verification

- [x] 4.1 `go build ./...` passes
- [x] 4.2 `go test ./...` passes — zero FAIL
