# fs_read Smart Reading + fs_stat Tool

## Summary

Add `Stat()` and `ReadWithMeta()` methods to the filesystem tool, and expose them as the `fs_stat` tool and enhanced `fs_read` tool with offset/limit parameters.

## Motivation

The current `fs_read` tool reads entire files. For large files, this wastes tokens and memory. Adding offset/limit support enables partial reads, and `fs_stat` enables metadata inspection without reading content.

## Non-goals

- Modifying `app.go` wiring (deferred to Unit 5)
- Changing existing `Read()` behavior (backward compatible)
