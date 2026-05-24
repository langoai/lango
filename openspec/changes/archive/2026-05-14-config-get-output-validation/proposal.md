## Why

`lango config get` uses `--output plain|json`, but docs and specs still reference `--format json`, and the implementation does not validate invalid output values before loading config.

## What Changes

- validate `lango config get --output` as `plain|json`
- reject unknown values before config loading
- add regression coverage
- sync docs and OpenSpec from stale `--format json` wording to the actual `--output json` contract

## Impact

- removes confusing config CLI documentation drift
- makes scripting behavior predictable
- avoids unnecessary config loading on invalid invocations
