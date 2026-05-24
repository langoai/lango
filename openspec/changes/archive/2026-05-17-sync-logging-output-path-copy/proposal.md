# Proposal: Sync Logging Output Path Copy

## Why

The logging default output now falls back to stderr when no explicit output path or writer is configured. The settings form still says an empty output path means stdout, which is incorrect and can mislead users configuring production logging.

## What Changes

- Update the Logging settings form placeholder and description to say empty output path uses stderr.
- Add a focused settings form test that prevents the stale stdout copy from returning.
- Update public configuration docs to include `logging.outputPath` and its stderr fallback.

## Impact

This is a copy and documentation sync for existing behavior. It does not change config parsing or logging runtime behavior.
