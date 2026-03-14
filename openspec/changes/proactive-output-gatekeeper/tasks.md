## 1. Content-Aware Compressor Library

- [x] 1.1 Create `internal/tooloutput/detect.go` with `DetectContentType` (JSON, Log, Code, StackTrace, Text)
- [x] 1.2 Create `internal/tooloutput/compress.go` with CompressJSON, CompressLog, CompressCode, CompressStackTrace, CompressHeadTail, Compress dispatcher
- [x] 1.3 Create `internal/tooloutput/compress_test.go` with table-driven tests for all compressors

## 2. Output Store

- [x] 2.1 Create `internal/tooloutput/store.go` with OutputStore (Store, Get, GetRange, Grep, lifecycle.Component)
- [x] 2.2 Create `internal/tooloutput/store_test.go` with TTL, range, grep tests
- [x] 2.3 Create `internal/app/tools_output.go` with `tool_output_get` tool (full/range/grep modes)

## 3. Output Manager Middleware

- [x] 3.1 Add `OutputManagerConfig` to `internal/config/types.go`
- [x] 3.2 Add viper defaults to `internal/config/loader.go`
- [x] 3.3 Create `internal/toolchain/mw_output_manager.go` with WithOutputManager using tooloutput.DetectContentType and tooloutput.Compress
- [x] 3.4 Create `internal/toolchain/mw_output_manager_test.go` with tier classification, store integration, and meta injection tests

## 4. Smart File Reading

- [x] 4.1 Add StatResult/ReadResult types and Stat/ReadWithMeta methods to `internal/tools/filesystem/filesystem.go`
- [x] 4.2 Add tests for Stat and ReadWithMeta in `internal/tools/filesystem/filesystem_test.go`
- [x] 4.3 Add `fs_stat` tool and offset/limit params to `fs_read` in `internal/app/tools_filesystem.go`

## 5. Wiring and Documentation

- [x] 5.1 Add `OutputStore` field to App struct in `internal/app/types.go`
- [x] 5.2 Replace WithTruncate with WithOutputManager in `internal/app/app.go`, wire OutputStore
- [x] 5.3 Register output category and tool_output_get in catalog
- [x] 5.4 Update `prompts/TOOL_USAGE.md` with tool_output_get, fs_stat, and offset/limit documentation
