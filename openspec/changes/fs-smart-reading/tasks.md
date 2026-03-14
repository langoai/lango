# Tasks

- [x] Add `StatResult` and `ReadResult` structs to `internal/tools/filesystem/filesystem.go`
- [x] Implement `Stat()` method with `bufio.Scanner` line counting
- [x] Implement `ReadWithMeta()` method with 1-indexed offset and limit support
- [x] Add `countLines()` helper function
- [x] Add table-driven tests for `Stat` (regular file, directory, non-existent)
- [x] Add table-driven tests for `ReadWithMeta` (full read, offset, limit, combined, beyond-EOF)
- [x] Update `fs_read` handler: add `offset`/`limit` params, use `toolparam.RequireString`
- [x] Add `fs_stat` tool definition in `buildFilesystemTools`
- [x] Migrate remaining handlers (`fs_write`, `fs_edit`, `fs_mkdir`, `fs_delete`) to `toolparam`
- [x] Verify build and all tests pass
