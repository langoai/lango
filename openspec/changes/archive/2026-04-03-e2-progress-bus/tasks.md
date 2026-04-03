## Tasks

### Task 1: Create ProgressEvent and ProgressType types
- **File**: `internal/streamx/progress.go`
- **Status**: DONE
- [x] Define `ProgressType` string type with constants: `ProgressStarted`, `ProgressUpdate`, `ProgressCompleted`, `ProgressFailed`
- [x] Define `ProgressEvent` struct with Source, Type, Message, Progress, Metadata fields

### Task 2: Implement ProgressBus with Subscribe/SubscribeAll/Emit
- **File**: `internal/streamx/progress.go`
- **Status**: DONE
- [x] Implement `ProgressBus` struct with `sync.RWMutex` and subscriber slice
- [x] Implement `NewProgressBus()` constructor
- [x] Implement `Subscribe(filter)` with buffered channel (cap 64) and cancel function
- [x] Implement `SubscribeAll()` as `Subscribe("")` wrapper
- [x] Implement `Emit(event)` with read lock and prefix-based filtering

### Task 3: Add non-blocking emit with buffer overflow handling
- **File**: `internal/streamx/progress.go`
- **Status**: DONE
- [x] Use `select/default` pattern for non-blocking channel send in Emit
- [x] Skip closed subscribers during emit iteration
- [x] Cancel function marks subscriber closed, closes channel, and removes from slice

### Task 4: Write tests (7 tests covering all scenarios)
- **File**: `internal/streamx/progress_test.go`
- **Status**: DONE
- [x] TestProgressBus_EmitAndSubscribe — prefix filtering works, non-matching events excluded
- [x] TestProgressBus_SubscribeAll — receives all events regardless of source
- [x] TestProgressBus_Cancel — channel closed, double-cancel safe, emit after cancel safe
- [x] TestProgressBus_BufferFullDropsEvent — events silently dropped when buffer full (64 cap)
- [x] TestProgressBus_ConcurrentEmit — concurrent goroutine safety
- [x] TestProgressBus_FilterPrefixMatching — bg: prefix matches bg:task-123 but not tool:bg_submit
- [x] TestProgressBus_MultipleSubscribers — multiple subscribers receive same event independently
